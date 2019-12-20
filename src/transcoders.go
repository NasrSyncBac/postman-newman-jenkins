package main

import (
	"time"

	"github.com/Syncbak-Git/elasticgo"
)

func ActiveSourceStreamCount() (int, error) {
	client, err := elasticgo.NewClient()
	check(err)

	start := time.Now().Add(-4 * time.Minute).UTC()
	end := time.Now().UTC()

	sourceStreamCount, err := client.SourceStreamCountAtTranscode(start, end)
	check(err)

	return int(sourceStreamCount), nil
}

func connectedTranscoders() (map[string][]string, error) {
	client, err := elasticgo.NewClientForCluster("qa")
	check(err)

	start := time.Now().Add(-4 * time.Minute).UTC()
	end := time.Now().UTC()
	transcoderClient, err := client.NewSearchClientBuilder().SearchRange(start, end).Filter("fields.name:Phase6Transcoder").Build()
	check(err)

	entries, err := transcoderClient.Entries()
	check(err)

	checkerMap := make(map[string][]string)
	for _, entry := range entries {
		serverName := entry.Fields.Host
		if _, found := checkerMap[serverName]; !found {
			checkerMap[serverName] = []string{"Various"}
		}
	}
	return checkerMap, nil
}

func transcoderWorkersInUse() (float64, error) {
	client, err := elasticgo.NewClient()
	check(err)
	start := time.Now().Add(-4 * time.Minute).UTC()
	end := time.Now().UTC()
	return client.TranscoderInProgressThreads(start, end)
}
