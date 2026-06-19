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

func insertESChildChunk(ctx context.Context, meta ChildChunkMeta, content string) error {
	cli, err := getESClient()
	if err != nil {
		return err
	}
	if err := ensureESIndex(ctx, cli); err != nil {
		return err
	}

	payload, err := json.Marshal(indexedChildDocument{
		ProjectID: meta.ProjectID,
		FileID:    meta.FileID,
		ChunkID:   meta.ChunkID,
		ParentIDs: meta.ParentIDs,
		Content:   content,
	})
	if err != nil {
		return err
	}

	res, err := cli.Index(
		esIndexName(),
		bytes.NewReader(payload),
		cli.Index.WithContext(ctx),
		cli.Index.WithDocumentID(chunkDocumentID(meta)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("index es child chunk failed: status=%s body=%s", res.Status(), strings.TrimSpace(string(body)))
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
