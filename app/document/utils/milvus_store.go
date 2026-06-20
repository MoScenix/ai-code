package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MoScenix/ai-code/app/document/conf"
	openaiemb "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

const (
	defaultEmbeddingBaseURL        = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultEmbeddingModelName      = "text-embedding-v1"
	defaultEmbeddingDimensions     = 1536
	defaultEmbeddingTimeoutSeconds = 60
	milvusIDField                  = "id"
	milvusContentField             = "content"
	milvusProjectIDField           = "project_id"
	milvusFileIDField              = "file_id"
	milvusChunkIDField             = "chunk_id"
	milvusParentIDsField           = "parent_ids"
	milvusVectorField              = "vector"
)

var (
	milvusClientOnce sync.Once
	milvusClient     *milvusclient.Client
	milvusClientErr  error
	milvusSchemaOnce sync.Once
	milvusSchemaErr  error
	embedderOnce     sync.Once
	qwenEmbedder     *openaiemb.Embedder
	qwenEmbedderErr  error
)

func insertMilvusChildChunks(ctx context.Context, children []ChildChunk, vectors [][]float32) error {
	if len(children) == 0 {
		return nil
	}
	if len(children) != len(vectors) {
		return fmt.Errorf("insert milvus child chunks got %d children and %d vectors", len(children), len(vectors))
	}
	cli, err := getMilvusClient(ctx)
	if err != nil {
		return err
	}
	if err := ensureMilvusCollection(ctx, cli); err != nil {
		return err
	}

	ids := make([]string, 0, len(children))
	contents := make([]string, 0, len(children))
	projectIDs := make([]int64, 0, len(children))
	fileIDs := make([]int64, 0, len(children))
	chunkIDs := make([]int64, 0, len(children))
	parentIDs := make([][]int64, 0, len(children))
	for _, child := range children {
		ids = append(ids, chunkDocumentID(child.Meta))
		contents = append(contents, child.Content)
		projectIDs = append(projectIDs, child.Meta.ProjectID)
		fileIDs = append(fileIDs, child.Meta.FileID)
		chunkIDs = append(chunkIDs, child.Meta.ChunkID)
		parentIDs = append(parentIDs, child.Meta.ParentIDs)
	}
	dim := 0
	if len(vectors) > 0 {
		dim = len(vectors[0])
	}

	_, err = cli.Insert(ctx, milvusclient.NewColumnBasedInsertOption(milvusCollectionName()).
		WithVarcharColumn(milvusIDField, ids).
		WithVarcharColumn(milvusContentField, contents).
		WithInt64Column(milvusProjectIDField, projectIDs).
		WithInt64Column(milvusFileIDField, fileIDs).
		WithInt64Column(milvusChunkIDField, chunkIDs).
		WithFloatVectorColumn(milvusVectorField, dim, vectors).
		WithColumns(column.NewColumnInt64Array(milvusParentIDsField, parentIDs)))
	if err != nil {
		return fmt.Errorf("insert milvus child chunks failed: %w", err)
	}
	return nil
}

func deleteMilvusProjectData(ctx context.Context, projectID int64) error {
	cli, err := getMilvusClient(ctx)
	if err != nil {
		return err
	}
	if err := ensureMilvusCollection(ctx, cli); err != nil {
		return err
	}
	collection := milvusCollectionName()
	expr := fmt.Sprintf(`%s == %d`, milvusProjectIDField, projectID)
	_, err = cli.Delete(ctx, milvusclient.NewDeleteOption(collection).WithExpr(expr))
	if err != nil {
		return fmt.Errorf("delete milvus project data failed: %w", err)
	}
	return nil
}

func searchMilvusChildren(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]RetrievedChild, error) {
	cli, err := getMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	if err := ensureMilvusCollection(ctx, cli); err != nil {
		return nil, err
	}
	vector, err := embedMilvusText(ctx, query)
	if err != nil {
		return nil, err
	}
	k := int(topK)
	if k <= 0 {
		k = 5
	}
	filter := fmt.Sprintf(`%s == %d && %s == %d`, milvusProjectIDField, projectID, milvusFileIDField, fileID)
	results, err := cli.Search(ctx, milvusclient.NewSearchOption(
		milvusCollectionName(),
		k,
		[]entity.Vector{entity.FloatVector(vector)},
	).WithANNSField(milvusVectorField).
		WithFilter(filter).
		WithOutputFields(milvusParentIDsField).
		WithConsistencyLevel(entity.ClStrong))
	if err != nil {
		return nil, fmt.Errorf("search milvus child chunks failed: %w", err)
	}

	children := make([]RetrievedChild, 0, k)
	for _, result := range results {
		if result.Err != nil {
			return nil, result.Err
		}
		parentIDsColumn := result.GetColumn(milvusParentIDsField)
		if parentIDsColumn == nil {
			continue
		}
		for i := 0; i < result.ResultCount; i++ {
			parentIDs, err := parentIDsFromColumn(parentIDsColumn, i)
			if err != nil {
				return nil, err
			}
			if len(parentIDs) == 0 {
				continue
			}
			children = append(children, RetrievedChild{
				ParentIDs: parentIDs,
			})
		}
	}
	return children, nil
}

func getMilvusClient(ctx context.Context) (*milvusclient.Client, error) {
	milvusClientOnce.Do(func() {
		cfg := conf.GetConf().Milvus
		address := strings.TrimSpace(cfg.Address)
		if address == "" {
			address = "127.0.0.1:19530"
		}
		milvusClient, milvusClientErr = milvusclient.New(ctx, &milvusclient.ClientConfig{
			Address:  address,
			Username: cfg.Username,
			Password: cfg.Password,
		})
	})
	return milvusClient, milvusClientErr
}

func ensureMilvusCollection(ctx context.Context, cli *milvusclient.Client) error {
	milvusSchemaOnce.Do(func() {
		collection := milvusCollectionName()
		exists, err := cli.HasCollection(ctx, milvusclient.NewHasCollectionOption(collection))
		if err != nil {
			milvusSchemaErr = err
			return
		}
		if exists {
			milvusSchemaErr = validateMilvusCollection(ctx, cli, collection)
			if milvusSchemaErr != nil {
				return
			}
			milvusSchemaErr = loadMilvusCollection(ctx, cli, collection)
			return
		}

		dim := conf.GetConf().Milvus.VectorDim
		if dim <= 0 {
			dim = defaultEmbeddingDimensions
		}
		schema := entity.NewSchema().
			WithField(entity.NewField().WithName(milvusIDField).WithDataType(entity.FieldTypeVarChar).WithMaxLength(128).WithIsPrimaryKey(true)).
			WithField(entity.NewField().WithName(milvusContentField).WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535)).
			WithField(entity.NewField().WithName(milvusProjectIDField).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(milvusFileIDField).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(milvusChunkIDField).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(milvusParentIDsField).WithDataType(entity.FieldTypeArray).WithElementType(entity.FieldTypeInt64).WithMaxCapacity(4)).
			WithField(entity.NewField().WithName(milvusVectorField).WithDataType(entity.FieldTypeFloatVector).WithDim(int64(dim)))

		indexOptions := []milvusclient.CreateIndexOption{
			milvusclient.NewCreateIndexOption(collection, milvusVectorField, index.NewHNSWIndex(entity.COSINE, 16, 200)).WithIndexName(milvusVectorField),
			milvusclient.NewCreateIndexOption(collection, milvusProjectIDField, index.NewSortedIndex()).WithIndexName(milvusProjectIDField),
			milvusclient.NewCreateIndexOption(collection, milvusFileIDField, index.NewSortedIndex()).WithIndexName(milvusFileIDField),
		}
		if err := cli.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collection, schema).WithIndexOptions(indexOptions...)); err != nil {
			milvusSchemaErr = fmt.Errorf("create milvus collection failed: %w", err)
			return
		}
		milvusSchemaErr = loadMilvusCollection(ctx, cli, collection)
	})
	return milvusSchemaErr
}

func validateMilvusCollection(ctx context.Context, cli *milvusclient.Client, collection string) error {
	coll, err := cli.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption(collection))
	if err != nil {
		return err
	}
	required := map[string]entity.FieldType{
		milvusIDField:        entity.FieldTypeVarChar,
		milvusContentField:   entity.FieldTypeVarChar,
		milvusProjectIDField: entity.FieldTypeInt64,
		milvusFileIDField:    entity.FieldTypeInt64,
		milvusChunkIDField:   entity.FieldTypeInt64,
		milvusParentIDsField: entity.FieldTypeArray,
		milvusVectorField:    entity.FieldTypeFloatVector,
	}
	actual := make(map[string]entity.FieldType, len(coll.Schema.Fields))
	for _, field := range coll.Schema.Fields {
		actual[field.Name] = field.DataType
	}
	for name, typ := range required {
		if actual[name] != typ {
			return fmt.Errorf("milvus collection %s has incompatible schema: field %s expects %v, got %v; recreate the collection to use scalar project/file fields", collection, name, typ, actual[name])
		}
	}
	return nil
}

func loadMilvusCollection(ctx context.Context, cli *milvusclient.Client, collection string) error {
	state, err := cli.GetLoadState(ctx, milvusclient.NewGetLoadStateOption(collection))
	if err == nil && state.State == entity.LoadStateLoaded {
		return nil
	}
	task, err := cli.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collection))
	if err != nil {
		return fmt.Errorf("load milvus collection failed: %w", err)
	}
	if err := task.Await(ctx); err != nil {
		return fmt.Errorf("await milvus collection load failed: %w", err)
	}
	return nil
}

func milvusCollectionName() string {
	collection := strings.TrimSpace(conf.GetConf().Milvus.Collection)
	if collection == "" {
		return "document_chunks"
	}
	return collection
}

func metadataInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

func metadataInt64Slice(value any) []int64 {
	switch v := value.(type) {
	case []int64:
		return v
	case []any:
		out := make([]int64, 0, len(v))
		for _, item := range v {
			if n := metadataInt64(item); n > 0 {
				out = append(out, n)
			}
		}
		return out
	case []float64:
		out := make([]int64, 0, len(v))
		for _, item := range v {
			if item > 0 {
				out = append(out, int64(item))
			}
		}
		return out
	default:
		return nil
	}
}

func parentIDsFromColumn(col column.Column, index int) ([]int64, error) {
	value, err := col.Get(index)
	if err != nil {
		return nil, err
	}
	return metadataInt64Slice(value), nil
}

func embedMilvusText(ctx context.Context, text string) ([]float32, error) {
	vectors, err := embedMilvusTexts(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 {
		return nil, fmt.Errorf("embed milvus text returned %d vectors", len(vectors))
	}
	return vectors[0], nil
}

func embedMilvusTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	emb, err := getQwenEmbedder(ctx)
	if err != nil {
		return nil, err
	}
	vectors, err := emb.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed milvus texts failed: %w", err)
	}
	if len(vectors) != len(texts) {
		return nil, fmt.Errorf("embed milvus texts returned %d vectors for %d texts", len(vectors), len(texts))
	}
	out := make([][]float32, 0, len(vectors))
	for _, vector := range vectors {
		row := make([]float32, len(vector))
		for i, value := range vector {
			row[i] = float32(value)
		}
		out = append(out, row)
	}
	return out, nil
}

func getQwenEmbedder(ctx context.Context) (*openaiemb.Embedder, error) {
	embedderOnce.Do(func() {
		dim := conf.GetConf().Milvus.VectorDim
		if dim <= 0 {
			dim = defaultEmbeddingDimensions
		}
		qwenEmbedder, qwenEmbedderErr = newQwenEmbedder(ctx, dim)
	})
	return qwenEmbedder, qwenEmbedderErr
}

func newQwenEmbedder(ctx context.Context, dimensions int) (*openaiemb.Embedder, error) {
	apiKey := strings.TrimSpace(os.Getenv("DASHSCOPE_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("DASHSCOPE_API_KEY is empty")
	}

	modelName := getEnv("EMBEDDING_MODEL_NAME", defaultEmbeddingModelName)
	baseURL := getEnv("DASHSCOPE_BASE_URL", defaultEmbeddingBaseURL)
	timeout := time.Duration(getEnvInt("EMBEDDING_TIMEOUT_SECONDS", defaultEmbeddingTimeoutSeconds)) * time.Second
	dim := getEnvInt("EMBEDDING_DIMENSIONS", dimensions)

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return openaiemb.NewEmbedder(ctx, &openaiemb.EmbeddingConfig{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		Model:      modelName,
		Dimensions: &dim,
	})
}

func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
