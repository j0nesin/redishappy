package sentinel

import (
	"fmt"
	"github.com/mdevilliers/redishappy/util"
	"sync"
	"time"
)

const (
	SentinelMarkedUp = iota
	SentinelMarkedDown = iota
	SentinelMarkedAlive = iota
)

type SentinelManager struct {
	eventsChannel chan SentinelEvent
	topologyRequestChannel chan TopologyRequest
}

var topologyState = SentinelTopology{Sentinels : map[string]*SentinelInfo{}}
var statelock = &sync.Mutex{}

func NewManager() *SentinelManager {
	events := make (chan SentinelEvent)
	requests := make (chan TopologyRequest)
	go loopEvents(events, requests)
	return &SentinelManager{ eventsChannel : events, topologyRequestChannel:requests}
}

func (m *SentinelManager) Notify(event SentinelEvent) {
	m.eventsChannel <- event
}

func (m *SentinelManager) GetState(request TopologyRequest) {
	m.topologyRequestChannel <- request
}

func (m *SentinelManager) ClearState() {
	statelock.Lock()
	defer statelock.Unlock()
	topologyState = SentinelTopology{Sentinels : map[string]*SentinelInfo{}}
}

func loopEvents(events chan SentinelEvent, topology chan TopologyRequest) {
 	for {
            select {
	            case event := <- events:
					updateState(event)
	            case read := <-topology:
	                read.ReplyChannel <- topologyState	            
            }
        }
}

func updateState(event interface{}) {

	statelock.Lock()
	defer statelock.Unlock()

	switch e := event.(type){
        case *SentinelAdded :

        	sentinel := e.GetSentinel()
        	info :=  &SentinelInfo{ SentinelLocation:sentinel.GetLocation(), 
									LastUpdated: time.Now().UTC(), 
									KnownClusters : []string{}, 
									State : SentinelMarkedUp }

			topologyState.Sentinels[ sentinel.GetLocation() ] = info

		case *SentinelLost :

			sentinel := e.GetSentinel()
			currentInfo, ok := topologyState.Sentinels[sentinel.GetLocation()]
			
			if ok {
				currentInfo.State = SentinelMarkedDown
				currentInfo.LastUpdated = time.Now().UTC()
			}

		case *SentinelPing :
			sentinel := e.GetSentinel()

			currentInfo, ok := topologyState.Sentinels[sentinel.GetLocation()]
			
			if ok {
				currentInfo.State = SentinelMarkedAlive
				currentInfo.LastUpdated = time.Now().UTC()
				currentInfo.KnownClusters = e.Clusters
			}			

        default:
           fmt.Println("Unknown sentinel event : ", util.String(e))
    }
}