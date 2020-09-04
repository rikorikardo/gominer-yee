package yee

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dchest/blake2b"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/sman2013/gominer-yee/clients"
)

// NewClient creates a new RpcClient given a '[stratum+tcp://]host:port' connectionstring
func NewClient(connectionstring, pooluser string) (sc clients.Client) {
	if strings.HasPrefix(connectionstring, "stratum+tcp://") {
		sc = &StratumClient{connectionstring: strings.TrimPrefix(connectionstring, "stratum+tcp://"), User: pooluser}
	} else {
		s := RpcClient{}
		s.yeeUrl = connectionstring
		sc = &s
	}
	return
}

// RpcClient is a simple client to a yee
type RpcClient struct {
	yeeUrl string
}

func decodeMessage(resp *http.Response) (msg string, err error) {
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var data struct {
		Message string `json:"message"`
	}
	if err = json.Unmarshal(buf, &data); err == nil {
		msg = data.Message
	}
	return
}

// Start does nothing
func (sc *RpcClient) Start() {}

// SetDeprecatedJobCall does nothing
func (sc *RpcClient) SetDeprecatedJobCall(call clients.DeprecatedJobCall) {}

// GetWork fetches new work from the SIA daemon
func (sc *RpcClient) GetWork() (target []byte, header []byte, deprecationChannel chan bool, job interface{}, err error) {
	// the deprecationChannel is not used but return a valid channel anyway
	deprecationChannel = make(chan bool)

	client := &http.Client{}
	body := bytes.NewBufferString("{\"id\":1, \"jsonrpc\":\"2.0\", \"method\":\"get_work\",\"params\":[]}")
	req, err := http.NewRequest("POST", sc.yeeUrl, body)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
	case 400:
		msg, errd := decodeMessage(resp)
		if errd != nil {
			err = fmt.Errorf("Status code %d", resp.StatusCode)
		} else {
			err = fmt.Errorf("Status code %d, message: %s", resp.StatusCode, msg)
		}
		return
	default:
		err = fmt.Errorf("Status code %d", resp.StatusCode)
		return
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	target, header, err = sc.decodeWork(buf)

	return
}

func (sc *RpcClient) decodeWork(buf []byte) (target []byte, header []byte, err error) {
	var data RpcResponse
	var merkle []byte

	if err = json.Unmarshal(buf, &data); err != nil {
		return
	}
	job := data.Result.(map[string]interface{})
	if merkle, err = hex.DecodeString(job["merkle_root"].(string)[2:]); err != nil {
		return
	}
	tar := job["target"].(string)[2:]
	for len(tar) < 64 {
		tar = "0" + tar
	}
	if target, err = hex.DecodeString(tar); err != nil {
		return
	}

	extra := [36]byte{0}
	check := [4]byte{159, 14, 68, 76}

	// Construct the work data
	header = make([]byte, 0, 80)
	header = append(header, merkle...)
	header = append(header, []byte{0, 0, 0, 0, 0, 0, 0, 0}[:]...)
	header = append(header, extra[:]...)
	header = append(header, check[:]...)

	return
}

// SubmitHeader reports a solved header to the YEE daemon
func (sc *RpcClient) Submit(header []byte, job interface{}) (err error) {
	h := blake2b.Sum256(header)
	log.Printf("Mined: %x", h)

	data := hex.EncodeToString(header[:76])

	body := bytes.NewBufferString("{\"id\":1, \"jsonrpc\":\"2.0\", \"method\":\"submit_work\",\"params\":[\"" + data + "\"]}")
	req, err := http.NewRequest("POST", sc.yeeUrl, body)
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	switch resp.StatusCode {
	case 204:
	default:
		msg, errd := decodeMessage(resp)
		if errd != nil {
			err = fmt.Errorf("Status code %d", resp.StatusCode)
		} else {
			err = fmt.Errorf("%s", msg)
		}
		return
	}
	return
}
