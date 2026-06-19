package utils

import (
	"context"
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
	"github.com/cloudwego/eino-ext/components/indexer/milvus2"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

const (
	defaultEmbeddingBaseURL        = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultEmbeddingModelName      = "text-embedding-v4"
	defaultEmbeddingDimensions     = 768
	defaultEmbeddingTimeoutSeconds = 60
)

var (
	milvusOnce sync.Once
	milvusIdx  *milvus2.Indexer
	milvusErr  error
)

func insertMilvusChildChunk(ctx context.Context, meta ChildChunkMeta, content string) error {
	idx, err := getMilvusIndexer(ctx)
	if err != nil {
		return err
	}

	_, err = idx.Store(ctx, []*schema.Document{
		{
			ID:      chunkDocumentID(meta),
			Content: content,
			MetaData: map[string]any{
				"projectId": meta.ProjectID,
				"fileId":    meta.FileID,
				"chunkId":   meta.ChunkID,
				"parentIds": meta.ParentIDs,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("index milvus child chunk failed: %w", err)
	}
	return nil
}

func deleteMilvusProjectData(ctx context.Context, projectID int64) error {
	return nil
}

func getMilvusIndexer(ctx context.Context) (*milvus2.Indexer, error) {
	milvusOnce.Do(func() {
		cfg := conf.GetConf().Milvus
		address := strings.TrimSpace(cfg.Address)
		if address == "" {
			address = "127.0.0.1:19530"
		}
		collection := strings.TrimSpace(cfg.Collection)
		if collection == "" {
			collection = "document_chunks"
		}

		dim := cfg.VectorDim
		if dim <= 0 {
			dim = defaultEmbeddingDimensions
		}

		emb, err := newQwenEmbedder(ctx, dim)
		if err != nil {
			milvusErr = err
			return
		}

		milvusIdx, milvusErr = milvus2.NewIndexer(ctx, &milvus2.IndexerConfig{
			ClientConfig: &milvusclient.ClientConfig{
				Address:  address,
				Username: cfg.Username,
				Password: cfg.Password,
			},
			Collection: collection,
			Vector: &milvus2.VectorConfig{
				Dimension:    int64(dim),
				MetricType:   milvus2.COSINE,
				IndexBuilder: milvus2.NewHNSWIndexBuilder().WithM(16).WithEfConstruction(200),
			},
			Embedding: emb,
		})
	})
	return milvusIdx, milvusErr
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
