package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	//golog "github.com/ipfs/go-log"
	//gologging "github.com/whyrusleeping/go-logging"
	peer "github.com/libp2p/go-libp2p-peer"
	//iaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/ChainSafeSystems/go-pre/p2p"
)

var LOCAL_PEER_ENDPOINT = "http://localhost:5001/api/v0/id"

type Simulator struct {
	nodes []*p2p.Service
}

// Borrowed from ipfs code to parse the results of the command `ipfs id`
type IdOutput struct {
	ID              string
	PublicKey       string
	Addresses       []string
	AgentVersion    string
	ProtocolVersion string
}

func NewSimulator(num int) (sim *Simulator, err error) {
	sim = new(Simulator)
	sim.nodes = make([]*p2p.Service, num)

	conf := &p2p.ServiceConfig{
		BootstrapNodes: []string{
			"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		},
		Port: 4000,
	}

	for i := 0; i < num; i++ {
		conf.Port = conf.Port + i
		sim.nodes[i] = new(p2p.Service)
		sim.nodes[i], err = p2p.NewService(conf)
		if err != nil {
			return nil, err
		}

	}

	return sim, nil
}

// quick and dirty function to get the local ipfs daemons address for bootstrapping
func getLocalPeerInfo() string {
	resp, err := http.Get(LOCAL_PEER_ENDPOINT)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var js IdOutput
	err = json.Unmarshal(body, &js)
	if err != nil {
		log.Fatalln(err)
	}
	for _, addr := range js.Addresses {
		// For some reason, possibly NAT traversal, we need to grab the loopback ip address
		if addr[0:8] == "/ip4/127" {
			return addr
		}
	}
	log.Fatalln(err)
	return ""
}

func main() {
	//golog.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info

	// num := 5

	// sim, err := NewSimulator(num)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for i := 0; i < num; i++ {
	// 	sim.nodes[i].Start()
	// }

	// var conf *p2p.ServiceConfig

	// if len(os.Args) < 2 {
	// 	conf = &p2p.ServiceConfig{
	// 		BootstrapNodes: []string{
	// 			getLocalPeerInfo(),
	// 		},
	// 		Port: 4000,
	// 	}
	// } else {
	// 	conf = &p2p.ServiceConfig{
	// 		BootstrapNodes: []string{
	// 			os.Args[1],
	// 		},
	// 		Port: 4000,
	// 	}
	// }

	conf := &p2p.ServiceConfig{
		BootstrapNodes: []string{
			getLocalPeerInfo(),
		},
		Port: 4000,
	}

	s, err := p2p.NewService(conf)
	if err != nil {
		log.Fatalf("NewService error: %s", err)
	}

	e := s.Start()
	if <-e != nil {
		log.Fatalf("Start error: %s", err)
	}

	// peer, err := iaddr.ParseString("/ip4/127.0.0.1/tcp/4000/ipfs/QmbtdfA4rPEkkub6YK3NijSGPaoxrYgPXMsWeg4Tu4zkxT")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// msg := []byte("hello")
	// err = s.Send(peer.ID(), msg)
	// if err != nil {
	// 	log.Fatalf("Send error: %s", err)
	// }

	if len(os.Args) < 2 {
		log.Println("listening")
		select{}
	}
	
	peerStr := os.Args[1]
	peerid, err := peer.IDB58Decode(peerStr[len(peerStr)-46:])
	if err != nil {
		log.Fatalln(err)
	}

	peer, err := s.DHT().FindPeer(s.Ctx(), peerid)
	if err != nil {
		log.Fatal(err)
	}

	swarm, err := s.NewSwarm()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s.Host().Peerstore().Peers())

	// open new stream with each peer
	stream, err := swarm.NewStream(s.Ctx(), s.Host().Peerstore().Peers()[0])
	if err != nil {
		log.Fatal(err)
	}
	stream.Write([]byte("Hello friend :)"))
	defer stream.Close()
	io.Copy(os.Stdout, stream) // pipe the stream to stdout

	// open new stream with each peer
	conn, err := swarm.DialPeer(s.Ctx(), peer.ID)
	if err != nil {
		log.Fatal(err)
	}

	stream, err = conn.NewStream()
	if err != nil {
		log.Fatal(err)
	}
	stream.Write([]byte("Hello friend :)"))
	defer stream.Close()
	io.Copy(os.Stdout, stream) // pipe the stream to stdout

	// stream, err := s.Host().NewStream(s.Ctx(), peer.ID, "/polkadot/0.0.0") 
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// _, err = stream.Write([]byte("Hello, world!\n"))
	// if err != nil {
	// 	log.Fatalln(err)
	// }

}