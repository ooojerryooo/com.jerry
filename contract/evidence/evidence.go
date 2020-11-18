package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"math/big"
	"time"
)

const (
	EVIDENCE = "Evidence"
	GRANT    = "Grant"
	LOG      = "OperateLog"
)

//存证对象
type Evidence struct {
	ObjectType string     `json:"objectType"`
	Header     *Header    `json:"header"`
	Body       string     `json:"body"`
	Signature  *Signature `json:"signature"`
}

//授权对象
type Grant struct {
	ObjectType            string `json:"objectType"`
	EvidenceCode          string `json:"evidenceCode"`          //"存证码"
	AuthorizedCertificate string `json:"authorizedCertificate"` //"证书"
	AuthorizedToken       string `json:"authorizedToken"`       //身份
	BeginTime             int64  `json:"beginTime"`             //开始时间,long类型
	EndTime               int64  `json:"endTime"`               //结束时间,long类型
	ReadTimes             int    `json:"readTimes"`             //取证次数,int类型
}

//操作日志
type OperateLog struct {
	ObjectType   string `json:"objectType"`
	EvidenceCode string `json:"evidenceCode"`
	OperateType  string `json:"operateType"`
	Operator     string `json:"operator"`
	Detail       string `json:"detail"`
}

type Header struct {
	EvidenceObjectCode string `json:"evidenceObjectCode"` //存证对象码
	Domain             string `json:"domain"`             //领域
	Application        string `json:"application"`        //应用
	DocumentType       string `json:"documentType"`       //单据类型
	TransactionType    string `json:"transactionType"`    //交易类型
	BizId              string `json:"bizId"`              //业务数据id
	EvidenceCode       string `json:"evidenceCode"`       //存证码
}

type Signature struct {
	Sign      string    `json:"sign"`
	Timestamp time.Time `json:"timestamp"`
}

type ECDSASignature struct {
	R, S *big.Int
}

type EvidenceCC struct {
}

func (v *EvidenceCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf("执行初始化方法!")
	return shim.Success(nil)
}

func (v *EvidenceCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()
	fmt.Printf("\n 方法: %s  参数 ： %s \n", fn, args)

	if fn == "set" {
		return v.set(stub, args)
	} else if fn == "get" {
		return v.get(stub, args)
	} else if fn == "grant" {
		return v.grant(stub, args)
	} else if fn == "searchEvidence" {
		return v.searchEvidence(stub, args)
	} else if fn == "queryLog" {
		return v.queryLog(stub, args)
	}

	return shim.Error("No this method:" + fn)
}

func (v *EvidenceCC) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	jsonData := args[0]
	var evidence Evidence

	err := json.Unmarshal([]byte(jsonData), &evidence)
	if err != nil {
		return shim.Error(fmt.Sprint("Failed to Unmarshal Evidence jsonData"))
	}

	evidence.ObjectType = EVIDENCE

	signp, _ := stub.GetSignedProposal()
	sign := hex.EncodeToString(signp.GetSignature())

	evidenceKey := evidence.Header.EvidenceCode
	evidence.Signature = &Signature{
		Sign:      sign,
		Timestamp: time.Now(),
	}

	evidenceJson, err := json.Marshal(evidence)

	err = stub.PutState(evidenceKey, evidenceJson)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to set evidence: %s", args[0]))
	}

	fmt.Println("save：", string(evidenceJson))

	fmt.Println("写日志")
	creator, _ := stub.GetCreator()
	log := &OperateLog{
		ObjectType:   LOG,
		EvidenceCode: evidenceKey,
		OperateType:  "put",
		Operator:     string(creator),
		Detail:       "存证上链",
	}
	err = writeLog(stub, log)
	if err != nil {
		return shim.Error(fmt.Sprint("Log write failure!"))
	}

	return shim.Success(evidenceJson)
}

func (v *EvidenceCC) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	evidenceCode := args[0]
	value, err := stub.GetState(evidenceCode)
	if err != nil {
		return shim.Error(fmt.Sprintf("There is no record of that Evidence %s!", evidenceCode))
	}
	fmt.Println("写日志")
	creator, _ := stub.GetCreator()
	log := &OperateLog{
		ObjectType:   LOG,
		EvidenceCode: evidenceCode,
		OperateType:  "get",
		Operator:     string(creator),
		Detail:       "根据存证ID获取链上数据",
	}
	err = writeLog(stub, log)
	if err != nil {
		return shim.Error(fmt.Sprint("Log write failure!"))
	}
	return shim.Success(value)
}

//查看某个存证信息的历史日志
func (v *EvidenceCC) queryLog(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	evidenceCode := args[0]
	iter, err := stub.GetHistoryForKey(LOG + "_" + evidenceCode)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed obtain %s Log!", evidenceCode))
	}
	defer iter.Close()
	var logs []OperateLog
	for iter.HasNext() {
		res, err := iter.Next()
		if err != nil {
			return shim.Error(fmt.Sprint("not Log History!"))
		}
		var log OperateLog
		_ = json.Unmarshal(res.Value, &log)
		logs = append(logs, log)
	}
	byteLogs, err := json.Marshal(logs)
	return shim.Success(byteLogs)
}

func (v *EvidenceCC) grant(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	jsonData := args[0]
	var grant Grant

	err := json.Unmarshal([]byte(jsonData), &grant)
	if err != nil {
		return shim.Error(fmt.Sprint("Failed to Unmarshal Grant jsonData"))
	}

	grantKey := fmt.Sprintf("%s_%s_%s", GRANT, grant.EvidenceCode, grant.AuthorizedToken)

	grantJson, _ := json.Marshal(grant)

	err = stub.PutState(grantKey, grantJson)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to set grant: %s", args[0]))
	}
	fmt.Println("saveGrant：", string(grantJson))

	fmt.Println("写日志")
	creator, _ := stub.GetCreator()
	log := &OperateLog{
		ObjectType:   LOG,
		EvidenceCode: grant.EvidenceCode,
		OperateType:  "grant",
		Operator:     string(creator),
		Detail:       "授权",
	}
	err = writeLog(stub, log)
	if err != nil {
		return shim.Error(fmt.Sprint("Log write failure!"))
	}

	return shim.Success(grantJson)
}

//查证
func (v *EvidenceCC) searchEvidence(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	evidenceKey := args[0]
	token := args[1]
	digest := sha256Hash(token)

	grantKey := fmt.Sprintf("%s_%s_%s", GRANT, evidenceKey, token)

	grantByte, err := stub.GetState(grantKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("There is no record of that Grant %s!", grantKey))
	}

	var grant Grant
	err = json.Unmarshal(grantByte, &grant)
	if err != nil {
		return shim.Error(fmt.Sprint("Failed to Unmarshal grantByte!"))
	}

	cert, err := byteToCert([]byte(grant.AuthorizedCertificate))
	if err != nil {
		return shim.Error(err.Error())
	}

	signature := args[2]
	rint, sint, err := getECDSASignatureRS(signature)
	if err != nil {
		return shim.Error(err.Error())
	}

	var flag bool
	switch cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		pub := cert.PublicKey.(*ecdsa.PublicKey)
		flag = ecdsa.Verify(pub, digest, rint, sint)
	case *rsa.PublicKey:
		publicKey := cert.PublicKey.(*rsa.PublicKey)
		byterun, err := hex.DecodeString(signature)
		err = rsa.VerifyPSS(publicKey, crypto.SHA256, digest, byterun, nil)
		if err != nil {
			fmt.Println("could not verify signature: ", err)
			return shim.Error(fmt.Sprint("could not verify signature: ", err))
		} else {
			flag = true
		}
	default:
		return shim.Error(fmt.Sprint("There is no certificate of this type!"))
	}

	fmt.Println("验签结果：", flag)
	if !flag {
		return shim.Error(fmt.Sprint("Failed to Verify signature!"))
	} else {

		if grant.ReadTimes < 1 {
			return shim.Error(fmt.Sprint("The number is not enough searchEvidence!"))
		}

		endTime := time.Unix(grant.EndTime/1000, 0)
		fmt.Printf("授权结束时间 = %+v", endTime)
		if endTime.Before(time.Now()) {
			return shim.Error(fmt.Sprint("SearchEvidence overtime!"))
		}

		//所有验证通过，获取存证
		evidence, err := stub.GetState(evidenceKey)
		if err != nil {
			return shim.Error(fmt.Sprintf("There is no record of that Evidence %s!", evidenceKey))
		}

		fmt.Printf("授权次数-1")
		grant.ReadTimes -= 1
		grantByte, _ = json.Marshal(grant)
		err = stub.PutState(grantKey, grantByte)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to set grant: %s", string(grantByte)))
		}

		fmt.Println("写日志")
		creator, _ := stub.GetCreator()
		log := &OperateLog{
			ObjectType:   LOG,
			EvidenceCode: grant.EvidenceCode,
			OperateType:  "searchEvidence",
			Operator:     string(creator),
			Detail:       "取证",
		}
		err = writeLog(stub, log)
		if err != nil {
			return shim.Error(fmt.Sprint("Log write failure!"))
		}

		return shim.Success(evidence)
	}
}

/**
  签名分解 通过hex解码，分割成数字证书r，s
*/
func getECDSASignatureRS(signature string) (rint, sint *big.Int, err error) {
	byterun, err := hex.DecodeString(signature)
	return UnmarshalECDSASignature(byterun)
}

func writeLog(stub shim.ChaincodeStubInterface, log *OperateLog) (err error) {
	logByte, err := json.Marshal(log)
	logKey := log.ObjectType + "_" + log.EvidenceCode
	err = stub.PutState(logKey, logByte)
	return
}

func getClientCert(stub shim.ChaincodeStubInterface) (cert *x509.Certificate, err error) {

	creatorByte, _ := stub.GetCreator()

	cert, err = byteToCert(creatorByte)

	fmt.Println("CertCN = " + cert.Subject.CommonName)

	return
}

func UnmarshalECDSASignature(raw []byte) (*big.Int, *big.Int, error) {
	// Unmarshal
	sig := new(ECDSASignature)
	_, err := asn1.Unmarshal(raw, sig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed unmashalling signature [%s]", err)
	}

	// Validate sig
	if sig.R == nil {
		return nil, nil, errors.New("invalid signature, R must be different from nil")
	}
	if sig.S == nil {
		return nil, nil, errors.New("invalid signature, S must be different from nil")
	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, R must be larger than zero")
	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, S must be larger than zero")
	}

	return sig.R, sig.S, nil
}

func sha256Hash(text string) []byte {
	SHA256Inst := sha256.New()
	SHA256Inst.Write([]byte(text))
	result := SHA256Inst.Sum(nil)
	return result
}

func byteToCert(b []byte) (cert *x509.Certificate, err error) {

	certStart := bytes.IndexAny(b, "-----BEGIN")
	if certStart == -1 {
		return nil, errors.New("No certificate found")
	}
	certText := b[certStart:]
	bl, _ := pem.Decode(certText)
	if bl == nil {
		return nil, errors.New("Could not decode the PEM structure")
	}

	cert, err = x509.ParseCertificate(bl.Bytes)
	if err != nil {
		return nil, errors.New("ParseCertificate failed")
	}

	return
}

func main() {
	err := shim.Start(new(EvidenceCC))
	if err != nil {
		fmt.Print(err.Error())
	}
}
