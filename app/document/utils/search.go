package utils

import (
	"context"
	"sort"
)

const rrfK = 60.0

type RetrievedChild struct {
	ParentIDs []int64
}

func SearchIndexedFile(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]int64, error) {
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
		return []int64{}, nil
	}
	if int64(len(rankedParents)) > topK {
		rankedParents = rankedParents[:topK]
	}

	parentIDs := make([]int64, 0, len(rankedParents))
	for _, parent := range rankedParents {
		parentIDs = append(parentIDs, parent.parentID)
	}
	return parentIDs, nil
}

func SearchByES(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]RetrievedChild, error) {
	return searchESChildren(ctx, projectID, fileID, query, topK)
}

func SearchByMilvus(ctx context.Context, projectID int64, fileID int64, query string, topK int64) ([]RetrievedChild, error) {
	return searchMilvusChildren(ctx, projectID, fileID, query, topK)
}

func DeleteProjectData(ctx context.Context, projectID int64) error {
	if err := deleteESProjectData(ctx, projectID); err != nil {
		return err
	}
	return deleteMilvusProjectData(ctx, projectID)
}

type parentRank struct {
	parentID int64
	score    float64
}

func fuseParentRanks(resultSets ...[]RetrievedChild) []parentRank {
	scores := map[int64]float64{}

	for _, results := range resultSets {
		for rank, child := range results {
			score := 1.0 / (rrfK + float64(rank+1))
			for _, parentID := range child.ParentIDs {
				if parentID <= 0 {
					continue
				}
				scores[parentID] += score
			}
		}
	}

	parents := make([]parentRank, 0, len(scores))
	for parentID, score := range scores {
		parents = append(parents, parentRank{
			parentID: parentID,
			score:    score,
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
