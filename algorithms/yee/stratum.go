package yee

import (
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/dchest/blake2b"
	"github.com/sman2013/gominer-yee/clients"
	"github.com/sman2013/gominer-yee/clients/stratum"
)

const (
	// HashSize is the length of a Yee hash
	HashSize = 32
)

// Target declares what a solution should be smaller than to be accepted
type Target [HashSize]byte

type stratumJob struct {
	JobID       string
	Merkle      []byte
	CustomExtra []byte
	CleanJobs   bool
	ExtraNonce2 stratum.ExtraNonce2
}

//StratumClient is a Yee client using the stratum protocol
type StratumClient struct {
	connectionstring string
	User             string

	mutex           sync.Mutex // protects following
	stratumclient   *stratum.Client
	sessionId       []byte
	extranonce2Size uint
	target          Target
	currentJob      stratumJob
	clients.BaseClient
}

//Start connects to the stratumserver and processes the notifications
func (sc *StratumClient) Start() {
	sc.mutex.Lock()
	defer func() {
		sc.mutex.Unlock()
	}()

	sc.DeprecateOutstandingJobs()

	sc.stratumclient = &stratum.Client{}
	//In case of an error, drop the current stratumclient and restart
	sc.stratumclient.ErrorCallback = func(err error) {
		log.Println("Error in connection to stratumserver:", err)
		sc.stratumclient.Close()
		sc.Start()
	}

	sc.subscribeToStratumDifficultyChanges()
	sc.subscribeToStratumJobNotifications()

	//Connect to the stratum server
	log.Println("Connecting to", sc.connectionstring)
	sc.stratumclient.Dial(sc.connectionstring)

	//Subscribe for mining
	//Close the connection on an error will cause the client to generate an error, resulting in te errorhandler to be triggered
	result, err := sc.stratumclient.Call("mining.subscribe", []string{"gominer"})
	if err != nil {
		log.Println("ERROR Error in response from stratum:", err)
		sc.stratumclient.Close()
		return
	}
	reply, ok := result.([]interface{})
	if !ok || len(reply) < 3 {
		log.Println("ERROR Invalid response from stratum:", result)
		sc.stratumclient.Close()
		return
	}

	// Keep the sessionId and extranonce2_size from the reply
	if sc.sessionId, err = stratum.HexStringToBytes(reply[1]); err != nil || len(sc.sessionId) != 4 {
		log.Println("ERROR Invalid sessionId from startum")
		sc.stratumclient.Close()
		return
	}

	extranonce2Size, ok := reply[2].(float64)
	if !ok {
		log.Println("ERROR Invalid extranonce2_size from stratum", reply[2], "type", reflect.TypeOf(reply[2]))
		sc.stratumclient.Close()
		return
	}
	sc.extranonce2Size = uint(extranonce2Size)

	//Authorize the miner
	go func() {
		result, err = sc.stratumclient.Call("mining.authorize", []string{sc.User, ""})
		if err != nil {
			log.Println("Unable to authorize:", err)
			sc.stratumclient.Close()
			return
		}
		log.Println("Authorization of", sc.User, ":", result)
	}()

}

func (sc *StratumClient) subscribeToStratumDifficultyChanges() {
	sc.stratumclient.SetNotificationHandler("mining.set_difficulty", func(params []interface{}) {
		if params == nil || len(params) < 1 {
			log.Println("ERROR No difficulty parameter supplied by stratum server")
			return
		}
		diff, ok := params[0].(float64)
		if !ok {
			log.Println("ERROR Invalid difficulty supplied by stratum server:", params[0])
			return
		}
		log.Println("Stratum server changed difficulty to", diff)
		sc.setDifficulty(diff)
	})
}

func (sc *StratumClient) subscribeToStratumJobNotifications() {
	sc.stratumclient.SetNotificationHandler("mining.notify", func(params []interface{}) {
		log.Println("New job received from stratum server")
		if params == nil || len(params) < 9 {
			log.Println("ERROR Wrong number of parameters supplied by stratum server")
			return
		}

		sj := stratumJob{}

		sj.ExtraNonce2.Size = sc.extranonce2Size

		var ok bool
		var err error
		if sj.JobID, ok = params[0].(string); !ok {
			log.Println("ERROR Wrong job_id parameter supplied by stratum server")
			return
		}

		// Convert the merkle parameter
		if sj.Merkle, err = stratum.HexStringToBytes(params[1]); err != nil {
			log.Println("ERROR Wrong merkle parameter supplied by stratum server")
			return
		}

		// Convert the custom extra parameter
		if sj.CustomExtra, err = stratum.HexStringToBytes(params[2]); err != nil {
			log.Println("Error Wrong extra parameter supplied by stratum server")
		}
		customExtraLen := 32 - int(sc.extranonce2Size)
		if len(sj.CustomExtra) > customExtraLen {
			sj.CustomExtra = sj.CustomExtra[0:customExtraLen]
		}

		if sj.CleanJobs, ok = params[8].(bool); !ok {
			log.Println("ERROR Wrong clean_jobs parameter supplied by stratum server")
			return
		}
		sc.addNewStratumJob(sj)
	})
}

func (sc *StratumClient) addNewStratumJob(sj stratumJob) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.currentJob = sj
	if sj.CleanJobs {
		sc.DeprecateOutstandingJobs()
	}
	sc.AddJobToDeprecate(sj.JobID)
}

// IntToTarget converts a big.Int to a Target.
func intToTarget(i *big.Int) (t Target, err error) {
	// Check for negatives.
	if i.Sign() < 0 {
		err = errors.New("Negative target")
		return
	}
	// In the event of overflow, return the maximum.
	if i.BitLen() > 256 {
		err = errors.New("Target is too high")
		return
	}
	b := i.Bytes()
	offset := len(t[:]) - len(b)
	copy(t[offset:], b)
	return
}

func difficultyToTarget(difficulty float64) (target Target, err error) {
	diffAsBig := big.NewFloat(difficulty)

	diffOneString := "0x00000000ffff0000000000000000000000000000000000000000000000000000"
	targetOneAsBigInt := &big.Int{}
	targetOneAsBigInt.SetString(diffOneString, 0)

	targetAsBigFloat := &big.Float{}
	targetAsBigFloat.SetInt(targetOneAsBigInt)
	targetAsBigFloat.Quo(targetAsBigFloat, diffAsBig)
	targetAsBigInt, _ := targetAsBigFloat.Int(nil)
	target, err = intToTarget(targetAsBigInt)
	return
}

func (sc *StratumClient) setDifficulty(difficulty float64) {
	target, err := difficultyToTarget(difficulty)
	if err != nil {
		log.Println("ERROR Error setting difficulty to ", difficulty)
	}
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.target = target
}

//GetWork fetches new work from the SIA daemon
func (sc *StratumClient) GetWork() (target, header []byte, deprecationChannel chan bool, job interface{}, err error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	job = sc.currentJob
	if sc.currentJob.JobID == "" {
		err = errors.New("No job received from stratum server yet")
		return
	}

	deprecationChannel = sc.GetDeprecationChannel(sc.currentJob.JobID)

	target = sc.target[:]

	var extra [36]byte
	copy(extra[:], sc.currentJob.CustomExtra) // custom extra
	copy(extra[32:], sc.sessionId)            // session id

	extraHash := blake2b.Sum256(extra[:])
	check := extraHash[:4]

	// Construct the work data
	header = make([]byte, 0, 80)
	header = append(header, sc.currentJob.Merkle...)
	header = append(header, []byte{0, 0, 0, 0, 0, 0, 0, 0}[:]...) //empty nonce
	header = append(header, extra[:]...)
	header = append(header, check...)

	return
}

//SubmitHeader reports a solution to the stratum server
func (sc *StratumClient) Submit(header []byte, job interface{}) (err error) {
	sj, _ := job.(stratumJob)
	nonce := hex.EncodeToString(header[32:40])
	extra := hex.EncodeToString(header[40:76])
	sc.mutex.Lock()
	c := sc.stratumclient
	sc.mutex.Unlock()
	stratumUser := sc.User
	if (time.Now().Nanosecond() % 100) == 0 {
		// do nothing
	}
	_, err = c.Call("mining.submit", []string{stratumUser, sj.JobID, extra, "", nonce})
	if err != nil {
		return
	}
	return
}
