package main

import (
	"crypto/sha256" //sha256哈希算法
	"encoding/hex"  //十六进制编码和解码
	"encoding/json" //json编码和解码
	"fmt"           //格式化输出
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"io"       //为 I/O 原语提供基本接口
	"log"      //Log 包实现了一个简单的日志包
	"net/http" //http包提供HTTP客户端和服务器实现
	"strconv"  //包strconv实现了对基本数据类型的字符串表示的转换
	"strings"  //打包字符串实现简单的函数来操纵 UTF-8 编码的字符串
	"sync"     //sync 提供基本的同步原语
	"time"     //提供了测量和显示时间的功能
)
const difficulty = 1
type Block struct{
	Index int
	Timestamp string
	Bike int
	Hash string
	PrevHash string
	Difficulty int
	Nonce string
}
var Blockchain []Block //存放区块数据
type Message struct{
	Bike int
}
var mutex = &sync.Mutex{}
//创建区块
func generateBlock(oldBlock Block,Bike int) Block {
	var newBlock Block
	t := time.Now()
	newBlock.Index = oldBlock.Index+1
	newBlock.Timestamp = t.String()
	newBlock.Bike = Bike
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty
	for i := 0;; i++ {
		hex := fmt.Sprintf("%x",i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), " do more work!")
			time.Sleep(time.Second)
			continue
		}else{
			fmt.Println(calculateHash(newBlock), " work done!")  //挖矿成功
			newBlock.Hash = calculateHash(newBlock)
			break
		}
	}
	return newBlock
}
func isHashValid(hash string, difficulty int) bool {
	//复制 difficulty 个0，并返回新字符串，当 difficulty 为 4 ，则 prefix 为 0000 
	prefix := strings.Repeat("0", difficulty)
	// 判断字符串 hash 是否包含前缀 prefix                            
	return strings.HasPrefix(hash, prefix)
}
func calculateHash(block Block) string {
	record:=strconv.Itoa(block.Index)+ block.Timestamp+strconv.Itoa(block.Bike)+block.PrevHash+block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false//确认Index的增长正确 
		}
	if oldBlock.Hash != newBlock.PrevHash {
		return false//确认PrevHash与前一个块的Hash相同   
		}
	if calculateHash(newBlock) != newBlock.Hash {
		return false
		}
	return true
}
func run() error { //run函数作为启动http服务器的函数
    mux := makeMuxRouter()//makeMuxRouter 主要定义路由处理
	//httpAddr := os.Getenv("ADDR")//.env            
	log.Println("Listening on ", "8888")
	s := &http.Server{
		Addr: ":" + "8888",
		Handler: mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		}
	if err := s.ListenAndServe(); err != nil {
		 return err
		 }
	return nil
}
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	//当收到GET请求，调用handleGetBlockchain函数
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	//当收到POST请求，调用handleWriteBlock函数
	return muxRouter
}
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	//处理HTTP的GET请求
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		 respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		 return
		}
	defer r.Body.Close()
	mutex.Lock()//产生区块
	newBlock := generateBlock(Blockchain[len(Blockchain)-1],m.Bike)
	mutex.Unlock()//判断区块的合法性
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) { //通过数组维护区块链
	Blockchain = append(Blockchain, newBlock)
	spew.Dump(Blockchain)
		}
	respondWithJSON(w, r, http.StatusCreated, newBlock)
}
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		 w.WriteHeader(http.StatusInternalServerError)
		 w.Write([]byte("HTTP 500: Internal Server Error"))
		return
		}
	w.WriteHeader(code)
	w.Write(response)
}
func main() {
	/*err := godotenv.Load()//允许我们读取.env
	if err != nil {
		log.Fatal(err)
		}*/
	go func() {
		t := time.Now()
		genesisBlock := Block{}//此乃创世区块
		genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), "", difficulty, ""}
		spew.Dump(genesisBlock)
		mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
		mutex.Unlock()
	}()
	log.Fatal(run())//启动web服务
}
