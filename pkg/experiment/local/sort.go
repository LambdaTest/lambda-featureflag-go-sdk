package local

import (
	"fmt"
	"github.com/LambdaTest/lambda-featureflag-go-sdk/internal/evaluation"
)

func topologicalSort(flags map[string]*evaluation.Flag, flagKeys []string) ([]*evaluation.Flag, error) {
	result := make([]*evaluation.Flag, 0)
	// Get the starting keys
	startingKeys := flagKeys
	if len(startingKeys) == 0 {
		startingKeys = make([]string, 0, len(flags))
		for k := range flags {
			startingKeys = append(startingKeys, k)
		}
	}
	// Track sorted flags instead of copying the flags map, so evaluating a few flags does not cost a copy of all of them
	visited := make(map[string]struct{}, len(startingKeys))
	// Sort into result
	for _, flagKey := range startingKeys {
		traversal, err := parentTraversal(flagKey, flags, visited, []string{})
		if err != nil {
			return nil, err
		}
		if len(traversal) > 0 {
			result = append(result, traversal...)
		}
	}
	return result, nil
}

func parentTraversal(flagKey string, flags map[string]*evaluation.Flag, visited map[string]struct{}, path []string) ([]*evaluation.Flag, error) {
	flag := flags[flagKey]
	if flag == nil {
		return nil, nil
	}
	if _, done := visited[flagKey]; done {
		return nil, nil
	}
	dependencies := flag.Dependencies
	if len(dependencies) == 0 {
		visited[flagKey] = struct{}{}
		return []*evaluation.Flag{flag}, nil
	}
	path = append(path, flagKey)
	result := make([]*evaluation.Flag, 0)
	for _, parentKey := range dependencies {
		if contains(path, parentKey) {
			return nil, fmt.Errorf("detected a cycle between flags %v", path)
		}
		traversal, err := parentTraversal(parentKey, flags, visited, path)
		if err != nil {
			return nil, err
		}
		if len(traversal) > 0 {
			result = append(result, traversal...)
		}
	}
	result = append(result, flag)
	visited[flagKey] = struct{}{}
	return result, nil
}
