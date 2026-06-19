package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const rrfK = 60.0

type RetrievedChild struct {
	FileID    int64
	ChunkID   int64
	ParentIDs []int64
}

type ParentSearchHit struct {
	FileID   int64
	ChunkID  int64
	ParentID int64
	Content  string
	Score    float64
}

func SearchIndexedFile(ctx context.Context, projectID int64, fileID int64, fileDir string, query string, topK int64) ([]ParentSearchHit, error) {
	if topK <= 0 {
		topK = 5
	}

	esChildren, err := SearchByES(ctx, projectID, fileID, query, topK)
	if err != nil {
		return nil, err
	}
	milvusChildren, err := SearchByMilvus(ctx, projectID, fileID, query, topK)
	if err != nil {
		return nil, err
	}

	rankedParents := fuseParentRanks(esChildren, milvusChildren)
	if len(rankedParents) == 0 {
		return []ParentSearchHit{}, nil
	}
	if int64(len(rankedParents)) > topK {
		rankedParents = rankedParents[:topK]
	}

	hits := make([]ParentSearchHit, 0, len(rankedParents))
	for _, parent := range rankedParents {
		content, err := readParentChunk(fileDir, parent.parentID)
		if err != nil {
			return nil, err
		}
		hits = append(hits, ParentSearchHit{
			FileID:   fileID,
			ChunkID:  parent.chunkID,
			ParentID: parent.parentID,
			Content:  content,
			Score:    parent.score,
		})
	}
	return hits, nil
}

func SearchByES(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]RetrievedChild, error) {
	return []RetrievedChild{}, nil
}

func SearchByMilvus(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]RetrievedChild, error) {
	return []RetrievedChild{}, nil
}

func DeleteProjectData(ctx context.Context, projectID int64) error {
	if err := deleteESProjectData(ctx, projectID); err != nil {
		return err
	}
	return deleteMilvusProjectData(ctx, projectID)
}

type parentRank struct {
	parentID int64
	chunkID  int64
	score    float64
}

func fuseParentRanks(resultSets ...[]RetrievedChild) []parentRank {
	type aggregate struct {
		chunkID int64
		score   float64
	}
	scores := map[int64]aggregate{}

	for _, results := range resultSets {
		for rank, child := range results {
			score := 1.0 / (rrfK + float64(rank+1))
			for _, parentID := range child.ParentIDs {
				if parentID <= 0 {
					continue
				}
				agg := scores[parentID]
				agg.score += score
				if agg.chunkID == 0 {
					agg.chunkID = child.ChunkID
				}
				scores[parentID] = agg
			}
		}
	}

	parents := make([]parentRank, 0, len(scores))
	for parentID, agg := range scores {
		parents = append(parents, parentRank{
			parentID: parentID,
			chunkID:  agg.chunkID,
			score:    agg.score,
		})
	}
	sort.SliceStable(parents, func(i, j int) bool {
		if parents[i].score == parents[j].score {
			return parents[i].parentID < parents[j].parentID
		}
		return parents[i].score > parents[j].score
	})
	return parents
}

func readParentChunk(fileDir string, parentID int64) (string, error) {
	path := filepath.Join(fileDir, "chunks", fmt.Sprintf("parent_%d.txt", parentID))
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
