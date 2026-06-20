package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/MoScenix/ai-code/app/document/conf"
	"github.com/elastic/go-elasticsearch/v8"
)

var (
	esOnce sync.Once
	esCli  *elasticsearch.Client
	esErr  error
)

type indexedChildDocument struct {
	ProjectID int64   `json:"projectId"`
	FileID    int64   `json:"fileId"`
	ChunkID   int64   `json:"chunkId"`
	ParentIDs []int64 `json:"parentIds"`
	Content   string  `json:"content"`
}

func insertESChildChunks(ctx context.Context, children []ChildChunk) error {
	if len(children) == 0 {
		return nil
	}
	cli, err := getESClient()
	if err != nil {
		return err
	}
	if err := ensureESIndex(ctx, cli); err != nil {
		return err
	}

	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	for _, child := range children {
		action := map[string]any{
			"index": map[string]any{
				"_id": chunkDocumentID(child.Meta),
			},
		}
		if err := encoder.Encode(action); err != nil {
			return err
		}
		if err := encoder.Encode(indexedChildDocument{
			ProjectID: child.Meta.ProjectID,
			FileID:    child.Meta.FileID,
			ChunkID:   child.Meta.ChunkID,
			ParentIDs: child.Meta.ParentIDs,
			Content:   child.Content,
		}); err != nil {
			return err
		}
	}

	res, err := cli.Bulk(
		bytes.NewReader(payload.Bytes()),
		cli.Bulk.WithContext(ctx),
		cli.Bulk.WithIndex(esIndexName()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("bulk index es child chunks failed: status=%s body=%s", res.Status(), strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Status int             `json:"status"`
			Error  json.RawMessage `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return err
	}
	if parsed.Errors {
		for _, item := range parsed.Items {
			for op, result := range item {
				if result.Status >= 300 {
					return fmt.Errorf("bulk index es child chunks failed: op=%s status=%d error=%s", op, result.Status, strings.TrimSpace(string(result.Error)))
				}
			}
		}
		return fmt.Errorf("bulk index es child chunks failed")
	}
	return nil
}

func deleteESProjectData(ctx context.Context, projectID int64) error {
	cli, err := getESClient()
	if err != nil {
		return err
	}
	if err := ensureESIndex(ctx, cli); err != nil {
		return err
	}

	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"projectId": projectID,
			},
		},
	}
	payload, err := json.Marshal(query)
	if err != nil {
		return err
	}

	res, err := cli.DeleteByQuery(
		[]string{esIndexName()},
		bytes.NewReader(payload),
		cli.DeleteByQuery.WithContext(ctx),
		cli.DeleteByQuery.WithConflicts("proceed"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("delete es project data failed: status=%s body=%s", res.Status(), strings.TrimSpace(string(body)))
	}
	return nil
}

func searchESChildren(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]RetrievedChild, error) {
	cli, err := getESClient()
	if err != nil {
		return nil, err
	}
	if err := ensureESIndex(ctx, cli); err != nil {
		return nil, err
	}
	if topK <= 0 {
		topK = 5
	}

	body := map[string]any{
		"size": topK,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{"term": map[string]any{"projectId": projectID}},
					{"term": map[string]any{"fileId": fileID}},
				},
				"must": []map[string]any{
					{"match": map[string]any{"content": query}},
				},
			},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	res, err := cli.Search(
		cli.Search.WithContext(ctx),
		cli.Search.WithIndex(esIndexName()),
		cli.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf("search es child chunks failed: status=%s body=%s", res.Status(), strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source indexedChildDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	children := make([]RetrievedChild, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		if len(hit.Source.ParentIDs) == 0 {
			continue
		}
		children = append(children, RetrievedChild{
			ParentIDs: hit.Source.ParentIDs,
		})
	}
	return children, nil
}

func refreshESIndex(ctx context.Context) error {
	cli, err := getESClient()
	if err != nil {
		return err
	}
	res, err := cli.Indices.Refresh(cli.Indices.Refresh.WithContext(ctx), cli.Indices.Refresh.WithIndex(esIndexName()))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("refresh es index failed: status=%s body=%s", res.Status(), strings.TrimSpace(string(body)))
	}
	return nil
}

func getESClient() (*elasticsearch.Client, error) {
	esOnce.Do(func() {
		cfg := conf.GetConf().ES
		addresses := cfg.Addresses
		if len(addresses) == 0 {
			addresses = []string{"http://127.0.0.1:9200"}
		}
		esCli, esErr = elasticsearch.NewClient(elasticsearch.Config{
			Addresses: addresses,
			Username:  cfg.Username,
			Password:  cfg.Password,
		})
	})
	return esCli, esErr
}

func ensureESIndex(ctx context.Context, cli *elasticsearch.Client) error {
	index := esIndexName()
	exists, err := cli.Indices.Exists([]string{index}, cli.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer exists.Body.Close()
	if exists.StatusCode == 200 {
		return nil
	}
	if exists.StatusCode != 404 {
		body, _ := io.ReadAll(io.LimitReader(exists.Body, 4096))
		return fmt.Errorf("check es index failed: status=%s body=%s", exists.Status(), strings.TrimSpace(string(body)))
	}

	mapping := `{
		"mappings": {
			"properties": {
				"projectId": {"type": "long"},
				"fileId": {"type": "long"},
				"chunkId": {"type": "long"},
				"parentIds": {"type": "long"},
				"content": {"type": "text"}
			}
		}
	}`
	res, err := cli.Indices.Create(index, cli.Indices.Create.WithContext(ctx), cli.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 400 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("create es index failed: status=%s body=%s", res.Status(), strings.TrimSpace(string(body)))
	}
	return nil
}

func esIndexName() string {
	if index := strings.TrimSpace(conf.GetConf().ES.Index); index != "" {
		return index
	}
	return "document_chunks"
}

func chunkDocumentID(meta ChildChunkMeta) string {
	return fmt.Sprintf("%d:%d:%d", meta.ProjectID, meta.FileID, meta.ChunkID)
}
