package controldb

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Syncbak-Git/logging"
	redigo "github.com/garyburd/redigo/redis"
)

//SourceStreamDb represents
type SourceStreamDb struct {
	StreamConnection *DbConnection
}

//Station represents station info: rawstreamid and callsign
type Station struct {
	RawStreamID string `json:"rawStreamID"`
	CallSign    string `json:"callSign"`
}

//IngesterHost has data showing SegmentIngest host info and their syncboxes.
type IngesterHost struct {
	IP       string    `json:"ip"`
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	Stations []Station `json:"stations"`
}

//DbConnection represents a connection to a redis database.
type DbConnection struct {
	Address        string
	Password       string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	connectTimeout time.Duration
}

//NewSourceStreamDb initializes a new SourceStreamDb pointer for querying source stream and catcher status.
func NewSourceStreamDb(address string, password string, timeout time.Duration) *SourceStreamDb {
	return &SourceStreamDb{StreamConnection: &DbConnection{Address: address, Password: password, readTimeout: timeout,
		writeTimeout: timeout, connectTimeout: timeout}}

}

const prefix string = "nameservice:stream:"

func connect(db *DbConnection) (redigo.Conn, error) {
	//conn, err := redigo.DialTimeout("tcp", db.Address, db.connectTimeout, db.readTimeout, db.writeTimeout)
	conn, err := redigo.Dial("tcp", db.Address,
		redigo.DialConnectTimeout(db.connectTimeout),
		redigo.DialReadTimeout(db.readTimeout),
		redigo.DialWriteTimeout(db.writeTimeout))

	if err != nil {
		return nil, err
	}

	if len(db.Password) > 0 {
		if _, err := conn.Do("AUTH", db.Password); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

//FetchSlotCount gets the total number of used slots across all catchers
func (db *SourceStreamDb) FetchSlotCount() (int, error) {
	conn, err := connect(db.StreamConnection)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	//step 1: get all the streams using keys commnad
	result, err := redigo.Strings(conn.Do("KEYS", prefix+"*"))
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

//FetchAllCatchers returns a map of all catchers and their associated source streams.  If a source stream exists
//but is not associated with a catcher in the db it will not be returned.
func (db *SourceStreamDb) FetchAllCatchers() (map[string][]string, error) {
	conn, err := connect(db.StreamConnection)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	//step 1: get all the streams using keys commnad
	result, err := redigo.Strings(conn.Do("KEYS", prefix+"*"))
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return make(map[string][]string), nil
	}
	conn.Send("MULTI")
	//send get for each key found nameservice:stream:sourcestreamid
	for _, resp := range result {
		conn.Send("GET", resp)
	}
	r, err := redigo.Strings(conn.Do("EXEC"))
	mapOfIPAddresses := make(map[string][]string)
	if err != nil {
		logging.L.Error(nil, "Error calling exec %s", err)
	}
	//add all responses to a map with the ip addr of host as key
	for i, ip := range r {
		if err != nil {
			logging.L.Error(nil, "error calling strings (receive) %s", err)
			return nil, err
		}
		mediaIDs, ok := mapOfIPAddresses[ip]
		mediaID := strings.TrimPrefix(result[i], prefix)
		if ok {
			mediaIDs = append(mediaIDs, mediaID)
		} else {
			//create entry with single slice
			mediaIDs = []string{mediaID}
		}
		mapOfIPAddresses[ip] = mediaIDs
	}
	conn.Send("MULTI")
	keys := []string{}
	for k := range mapOfIPAddresses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	//get all the hostlookup values, which provide hostnames
	for _, key := range keys {
		conn.Send("GET", "hostlookup:"+key)
	}

	//also determine the activehost for each host pair (primary, backup)
	for _, key := range keys {
		val := mapOfIPAddresses[key]
		for _, id := range val {
			conn.Send("GET", "nameservice:activehost:"+id)
		}
	}
	names, err := redigo.Strings(conn.Do("EXEC"))
	if err != nil {
		return nil, err
	}
	index := 0
	hosts := make(map[string]string)
	//gmk 5/13/2016 - This is god-awful confusing - If we had this data in some relational or object data store, it would be much simpler!
	conn.Send("MULTI")
	//retrieve all the hosttype values
	for i, key := range keys {
		hosts[key] = names[i]
		conn.Send("GET", "hosttype:"+names[i])
		index++
	}
	//final exec to get the 'type' of host - this apparently is its purpose i.e. 720p
	types, err := redigo.Strings(conn.Do("EXEC"))
	if err != nil {
		return nil, err
	}
	ingesterTypes := make(map[string]string)
	for i, key := range keys {
		ids, ok := mapOfIPAddresses[key]
		ingesterTypes[key] = types[i]
		if !ok {
			logging.L.Error(nil, "did not find activehost key %s", key)
			continue
		}
		for i := range ids {
			hostName := names[index]
			//concatenate the host name and the source stream id of the station
			ids[i] = hostName + ":" + ids[i]
			index++
		}
		//get the old key, which should be the mediaids
		mapOfIPAddresses[key] = ids
	}
	final := make(map[string][]string)
	typeIndex := 0
	for k, v := range mapOfIPAddresses {
		hostName, ok := hosts[k]
		if !ok {
			logging.L.Error(nil, "did not have the key %s in hosts map!", k)
		} else {
			if len(hostName) <= 1 {
				hostName = "unknown"
			}
			iType, ok := ingesterTypes[k]
			if !ok {
				logging.L.Error(nil, "did not have the key %s in ingester types map!", k)
			} else {
				if len(iType) <= 1 {
					iType = "unknown"
				}
				final[hostName+" : "+k+" : "+iType] = v
			}
		}
		typeIndex++
	}
	return final, nil
}

//FetchIngesterHosts returns an array of IngesterHost structs.
func (db *SourceStreamDb) FetchIngesterHosts() ([]*IngesterHost, error) {
	cMap, err := db.FetchAllCatchers()
	if err != nil {
		return nil, err
	}
	var arr []*IngesterHost
	for k, v := range cMap {
		ih := IngesterHost{}
		ipName := strings.Split(k, ":")
		if len(ipName) != 3 {
			return nil, fmt.Errorf("The key %s should have had one colon!", k)
		}
		ih.Name = strings.TrimSpace(ipName[0])
		ih.IP = strings.TrimSpace(ipName[1])
		ih.Type = strings.TrimSpace(ipName[2])
		var stations []Station
		for _, s := range v {
			station := parseStation(strings.TrimSpace(strings.Replace(s, ".syncbak.corp", "", 2)))
			stations = append(stations, station)
		}
		ih.Stations = stations
		arr = append(arr, &ih)
	}
	return arr, nil
}

func parseStation(s string) Station {
	ar := strings.Split(s, ":")
	station := Station{}
	station.CallSign = ar[0]
	if len(ar) >= 2 {
		station.RawStreamID = ar[1]
	}
	return station
}
