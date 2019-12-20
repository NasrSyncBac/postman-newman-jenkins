package main

import (
	"sort"
	"time"

	"github.com/Syncbak-Git/elasticgo"
)

type Assignment struct {
	Queue string `json:"queue"`
	Host  string `json:"host"`
}

func adapterAssignments() (map[string][]string, error) {
	hosts := make(map[string][]string)

	f, err := elasticgo.NewFinderForCluster("qa", time.Now().Add(-10*time.Minute), time.Now())
	check(err)
	f.Client.MaxResults = 2000
	res, err := f.Name("cdnadapter").Find()
	check(err)

	sort.Slice(res.Entries, func(i, j int) bool {
		return res.Entries[i].Timestamp.After(res.Entries[j].Timestamp)
	})
	em := make(map[string]elasticgo.Entry)
	for _, b := range res.Entries {
		_, ok := em[b.Fields.SourceStreamID]
		if !ok {
			em[b.Fields.SourceStreamID] = b
		}
	}

	for ssid, entry := range em {
		queues, ok := hosts[entry.Fields.IP]
		if !ok {
			hosts[entry.Fields.IP] = []string{ssid}
		} else {
			queues = append(queues, ssid)
			hosts[entry.Fields.IP] = queues
		}
	}
	return hosts, nil
}
