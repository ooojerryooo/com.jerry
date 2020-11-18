package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"testing"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", res.Message)
		t.FailNow()
	}
}

func TestEvidenceCC_Init(t *testing.T) {
	scc := new(EvidenceCC)
	stub := shim.NewMockStub("evidence", scc)
	checkInit(t, stub, [][]byte{[]byte("init"), []byte("A"), []byte("123"), []byte("B"), []byte("234")})
}

func TestEvidenceCC_Invoke(t *testing.T) {
	scc := new(EvidenceCC)
	stub := shim.NewMockStub("evidence", scc)
	var value = `{
  "header": {
    "evidenceObjectCode": "e-contract",
    "domain": "",
    "application": "",
    "documentType": "",
    "transactionType": "",
    "bizId": "65e4195245804e2183f26b32c495d026",
    "evidenceCode": "1326069383327514624"
  },
  "body": "{\"code\":\"65e4195245804e2183f26b32c495d026\",\"owner\":\"f0krpqfu\",\"operator\":\"8a5d960b-36d8-4e9f-a904-8027274e59bb\",\"contractFileHash\":\"ae6e08933cc8212e33d902a81e50996f0985b69171f4475b944f5d4c127b7497\",\"contractAnnexHash\":[\"5cbced72c693e012dd9ca9096135f29c0f26153ce4055aef26fec12b6e77b94f\"],\"signList\":[{\"signature\":\"AAA\",\"timeStamp\":\"D:20200831193842+08'00'\",\"personInfo\":{\"id\":\"be270825-ed3d-4a07-98e5-1e6b2b28d5d7\",\"name\":\"黄仁华\",\"certType\":0,\"certNum\":\"500233199303268854\"}},{\"signature\":\"AA\",\"timeStamp\":\"D:20200831194030+08'00'\",\"orgInfo\":{\"id\":\"1804453094134016\",\"name\":\"用友网络科技股份有限公司\",\"unifiedSocialCreditCode\":\"91110000600001760P\"}},{\"signature\":\"AAAAA\",\"timeStamp\":\"D:20200831194030+08'00'\",\"orgInfo\":{\"id\":\"1804453094134016\",\"name\":\"用友网络科技股份有限公司\",\"unifiedSocialCreditCode\":\"91110000600001760P\"}},{\"signature\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\",\"timeStamp\":\"D:20200831194030+08'00'\",\"orgInfo\":{\"id\":\"1804453094134016\",\"name\":\"用友网络科技股份有限公司\",\"unifiedSocialCreditCode\":\"91110000600001760P\"}}],\"domainCode\":\"e-contract\"}"
}`
	res := stub.MockInvoke("1", [][]byte{[]byte("set"), []byte(value)})
	fmt.Println("上链结果" + res.String())

	res = stub.MockInvoke("1", [][]byte{[]byte("get"), []byte("1326069383327514624")})
	fmt.Println("查询结果" + res.String())

	var grant = `{
  "authorizedToken": "123",
  "authorizedCertificate": "-----BEGIN CERTIFICATE-----\nMIICLDCCAdKgAwIBAgIQOjPdh27dOxJSRR0DqAeIbzAKBggqhkjOPQQDAjBsMQsw\nCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy\nYW5jaXNjbzEUMBIGA1UEChMLZXhhbXBsZS5jb20xGjAYBgNVBAMTEXRsc2NhLmV4\nYW1wbGUuY29tMB4XDTIwMTEwMzA4MzkwMFoXDTMwMTEwMTA4MzkwMFowVjELMAkG\nA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBGcmFu\nY2lzY28xGjAYBgNVBAMMEUFkbWluQGV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYI\nKoZIzj0DAQcDQgAEXpMfJac3o6DtjMBqNEYCK1wqTI1Kt4Ep4Wx6HCbijaJ9LDBV\nMPCtxGQjCXOdewylqNoz64bJN2h2nzBO9nKJHKNsMGowDgYDVR0PAQH/BAQDAgWg\nMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMCsG\nA1UdIwQkMCKAIPg5q7ZlNPOhtfm+hhkUBfcSGJgiL7lON2BasbRxR8KBMAoGCCqG\nSM49BAMCA0gAMEUCIQD/5H/OO2zQyHvGBQurZbKbKf+XNgvWgSwG/nV+iVp+QwIg\nS+d0oEJPBC1asVe78VDdPNghGteU65h7/HUYhZdupUY=\n-----END CERTIFICATE-----",
  "evidenceCode": "1326069383327514624",
  "beginTime": 1608291770000,
  "endTime": 1610538170000,
  "readTimes": 1
}`
	res = stub.MockInvoke("1", [][]byte{[]byte("grant"), []byte(grant)})
	fmt.Println("授权结果" + res.String())

	res = stub.MockInvoke("1", [][]byte{[]byte("searchEvidence"), []byte("1326069383327514624"), []byte("123"), []byte("3046022100f1a0342dae9f8feb5902f5ae9cf5101958a439c59117dba41fd3dcab653fa807022100af3abffd83f604c031d40ed635c4ea826b927758041ac0a174d046154863a1cc")})
	fmt.Println("查证结果" + res.String())

	res = stub.MockInvoke("1", [][]byte{[]byte("searchEvidence"), []byte("1326069383327514624"), []byte("123"), []byte("3046022100f1a0342dae9f8feb5902f5ae9cf5101958a439c59117dba41fd3dcab653fa807022100af3abffd83f604c031d40ed635c4ea826b927758041ac0a174d046154863a1cc")})
	fmt.Println("查证结果" + res.String())
}
