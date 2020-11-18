package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"io"
	"strconv"
	"strings"
	"time"
)

type hwChaincode struct {
}

const (
	dateFomat                      = "2006-01-02 15:04:05"             //日期格式
	trace_suf                      = "trace"                           //溯源结构
	combinedConstruction_M         = "Order~Buyer~VaccId"              //订单联合主键买家+疫苗编码
	combinedConstruction_N         = "Order~Seller~VaccId"             //订单联合主键卖家+疫苗编码
	combinedConstruction_V         = "Vaccine~orgid~checkState~vaccid" //疫苗联合主键机构+检验状态+编码
	combinedConstruction_Inventory = "Vaccine~orgid~vaccid"            //库存联合主键 机构+编码
)

/*******结构体*******/
//企业信息
type OrgInfo struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	Address          string `json:"address"`
	PermissionNumber string `json:"permissionNumber"`
}

//疫苗信息
type Vaccine struct {
	Id         string `json:"id"`         //记录编码：规则结构体名+疫苗编码，唯一键
	VaccId     string `json:"vaccid"`     //疫苗编码：0692100292323
	Name       string `json:"name"`       //疫苗名称：白破疫苗
	Spec       string `json:"spec"`       //规格：1ml
	Count      string `json:"count"`      //数量：100
	Date       string `json:"date"`       //生产日期：2019-01-02 15:04:05
	Site       string `json:"site"`       //生产地址：天津市西青区工业大学软件学院
	Producer   string `json:"producer"`   //生产商：海王医疗
	Owner      string `json:"owner"`      //当前所有者：海王销售
	CheckState string `json:"checkState"` //检验状态：0未检验1合格2不合格
	CheckOrg   string `json:"checkOrg"`   //检验机构
	State      int    `json:"state"`      //0生产1销售2售出3检验合格
}

//订单信息
type OrderForm struct {
	OrderId     string   `json:"orderId"`     //订单Id
	VaccineIds  []string `json:"vaccineIds"`  //疫苗Id
	VaccNum     string   `json:"vaccNum"`     //采购数量
	VaccineName string   `json:"vaccineName"` //疫苗名称
	Spec        string   `json:"spec"`        //规格：1ml
	Site        string   `json:"site"`        //发货地址
	Date        string   `json:"date"`        //采购日期
	CheckState  string   `json:"checkState"`  //检验状态：检验合格
	BuyerId     string   `json:"buyerId"`     //买家Id
	BuyerName   string   `json:"buyerName"`   //买方公司名
	BuyerSite   string   `json:"buyerSite"`   //买家地址
	SellerId    string   `json:"sellerId"`    //卖家Id
	SellerName  string   `json:"sellerName"`  //卖方公司名
	SellerSite  string   `json:"sellerSite"`  //卖家地址
	Producer    string   `json:"producer"`    //生产商：海王医疗
	Owner       string   `json:"owner"`       //当前所有者：海王销售
	State       int      `json:"state"`       //0未背书1背书中2背书成功3背书失败
}

//销售信息
type SellInfo struct {
	Id          string `json:"id"`          //订单Id
	VaccId      string `json:"vaccId"`      //疫苗编码
	VaccineName string `json:"vaccineName"` //疫苗名称
	VaccNum     string `json:"vaccNum"`     //采购数量
	BuyerName   string `json:"buyerName"`   //消费者
	Spec        string `json:"spec"`        //规格：1ml
	Site        string `json:"site"`        //售卖网点
	SellerId    string `json:"sellerId"`    //卖家Id
	SellerName  string `json:"sellerName"`  //卖方公司名
	SellerSite  string `json:"sellerSite"`  //卖家地址
	Date        string `json:"date"`        //销售时间
	CheckState  string `json:"checkState"`  //终端售出
}

//溯源结构
type Trace struct {
	Id     string `json:"id"`     //疫苗Id：001，存储结构+trace
	Name   string `json:"name"`   //疫苗名称：白破疫苗
	Date   string `json:"date"`   //发生时间
	Site   string `json:"site"`   //位置
	Owner  string `json:"owner"`  //所有人
	Action string `json:"action"` //动作
	Result string `json:"result"` //结果
}

//评价结构
type Evaluate struct {
	Id           string `json:"id"`
	User         string `json:"user"`
	Phone        string `json:"phone"`
	VaccId       string `json:"vaccId"`       //疫苗编码
	EvaBaseInfo  string `json:"evaBaseInfo"`  //基本信息
	EvaSellDate  string `json:"evaSellDate"`  //购买时间
	EvaSellSite  string `json:"evaSellSite"`  //购买地点
	EvaSellPlace string `json:"evaSellPlace"` //购买场所
	EvaSellState string `json:"evaSellState"` //购买状态
	Complain     string `json:"complain"`     //投诉信息
}

func (t *hwChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()
	fmt.Println("===========================================================================")
	fmt.Printf("\n 方法: %s  参数 ： %s \n", fn, args)

	if fn == "addOrg" {
		return t.addOrg(stub, args)
	} else if fn == "productionVaccine" {
		return t.productionVaccine(stub, args)
	} else if fn == "queryAllVaccineList" {
		return t.queryAllVaccineList(stub, args)
	} else if fn == "checkVaccine" {
		return t.checkVaccine(stub, args)
	} else if fn == "getCheckedVaccIdList" {
		return t.getCheckedVaccIdList(stub, args)
	} else if fn == "addOrderA" {
		return t.addOrderA(stub, args)
	} else if fn == "queryInventory" {
		return t.queryInventory(stub, args)
	} else if fn == "addOrderB" {
		return t.addOrderB(stub, args)
	} else if fn == "SendEndorse" {
		return t.SendEndorse(stub, args)
	} else if fn == "comfirmEndorse" {
		return t.comfirmEndorse(stub, args)
	} else if fn == "addSellInfo" {
		return t.addSellInfo(stub, args)
	} else if fn == "submitEvaluate" {
		return t.submitEvaluate(stub, args)
	} else if fn == "queryAllSellInfoList" {
		return t.queryAllSellInfoList(stub, args)
	} else if fn == "queryAllOrderList" {
		return t.queryAllOrderList(stub, args)
	} else if fn == "queryOrderForm" {
		return t.queryOrderForm(stub, args)
	} else if fn == "queryUnEndorseOrder" {
		return t.queryUnEndorseOrder(stub, args)
	} else if fn == "updateOrderB" {
		return t.updateOrderB(stub, args)
	} else if fn == "queryTraceHistory" {
		return t.queryTraceHistory(stub, args)
	}

	return shim.Error("No this method:" + fn)
}

/*******合约核心方法区*******/
//添加企业（后台添加）
func (t *hwChaincode) addOrg(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var org OrgInfo
	org.Id = args[0]
	org.Name = args[1]
	org.Address = args[2]
	org.PermissionNumber = args[3]
	orgJson, _ := json.Marshal(org)
	stub.PutState(org.Id, orgJson)
	fmt.Println("企业: " + string(orgJson))
	return shim.Success(orgJson)
}

//生产疫苗productionVaccine.html
func (t *hwChaincode) productionVaccine(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	orgId := args[0]
	vaccStrs := args[1]
	var vaccines []Vaccine
	if err := json.Unmarshal([]byte(vaccStrs), &vaccines); err != nil {
		return shim.Error("json反序列化失败")
	}
	for i, vacc := range vaccines {
		vaccc, _ := t.getVaccineById(stub, "Vaccine"+vacc.VaccId)
		if vaccc.VaccId == vacc.VaccId {
			return shim.Error("vaccid exist:" + vacc.VaccId)
		}
		if vacc.VaccId != "" {
			vacc.Id = "Vaccine" + vacc.VaccId
			vacc.State = 0
			vacc.Owner = orgId
			vacc.Date = vacc.Date
			vacc.Producer = orgId
			vacc.CheckState = "0"
			vaccJson, _ := json.Marshal(vacc)
			stub.PutState(vacc.Id, []byte(vaccJson))
			fmt.Println("疫苗: " + string(vaccJson))

			//加库存
			indexKey, _ := stub.CreateCompositeKey(combinedConstruction_Inventory, []string{orgId, vacc.VaccId})
			stub.PutState(indexKey, []byte(vaccJson))

			var trace Trace
			trace.Id = vacc.VaccId + trace_suf
			trace.Name = vacc.Name
			trace.Date = vacc.Date
			trace.Site = vacc.Site
			trace.Owner = vacc.Owner
			trace.Action = "生产"
			trace.Result = "合格一致"
			traceStr, _ := json.Marshal(trace)
			stub.PutState(trace.Id, []byte(traceStr))
			fmt.Println("追溯: " + string(traceStr))
		} else {
			fmt.Println("第" + strconv.Itoa(i) + "行无编码！")
		}
	}
	return shim.Success([]byte("success"))
}

//检验疫苗
func (t *hwChaincode) checkVaccine(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	id := args[0]       //疫苗编码
	checkOrg := args[1] //检验机构
	checkState := args[2]
	vacc, _ := t.getVaccineById(stub, id)
	vacc.CheckState = checkState //1合格2不合格
	vacc.CheckOrg = checkOrg
	vaccJson, _ := json.Marshal(vacc)
	stub.PutState(id, []byte(vaccJson))
	fmt.Println("疫苗: " + string(vaccJson))

	//修改库存
	indexKey1, _ := stub.CreateCompositeKey(combinedConstruction_Inventory, []string{vacc.Producer, vacc.VaccId})
	stub.PutState(indexKey1, []byte(vaccJson))

	//下拉选框使用
	indexKey, _ := stub.CreateCompositeKey(combinedConstruction_V, []string{vacc.Owner, checkState, vacc.VaccId})
	stub.PutState(indexKey, []byte(vaccJson))

	var cstZone = time.FixedZone("CST", 8*3600)

	var trace Trace
	trace.Id = vacc.VaccId + trace_suf
	trace.Name = vacc.Name
	trace.Date = time.Now().In(cstZone).Format(dateFomat)
	trace.Site = vacc.Site
	trace.Owner = vacc.Owner
	if checkState == "1" {
		trace.Action = "检验合格"
		trace.Result = "合格一致"
	} else {
		trace.Action = "检验不合格"
		trace.Result = "不一致"
	}
	traceStr, _ := json.Marshal(trace)
	stub.PutState(trace.Id, []byte(traceStr))
	fmt.Println("追溯: " + string(traceStr))

	return shim.Success([]byte("success"))
}

//获取生产商疫苗检验合格的下拉列表
func (t *hwChaincode) getCheckedVaccIdList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	orgId := args[0]
	resultIterator, err := stub.GetStateByPartialCompositeKey(combinedConstruction_V, []string{orgId, "1"})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultIterator.Close()
	var vaccines []Vaccine
	for resultIterator.HasNext() {
		item, err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var vaccine Vaccine
		if err := json.Unmarshal(item.Value, &vaccine); err != nil {
			return shim.Error("json反序列化失败")
		}
		vaccines = append(vaccines, vaccine)
	}
	vaccsJson, _ := json.Marshal(vaccines)
	fmt.Println("下拉列表: ", string(vaccsJson))
	return shim.Success(vaccsJson)
}

//生产者添加订单addOrderA.html
func (t *hwChaincode) addOrderA(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	orgId := args[0]
	orderStrs := args[1]
	var orders []OrderForm
	if err := json.Unmarshal([]byte(orderStrs), &orders); err != nil {
		return shim.Error("json反序列化失败")
	}
	seller, _ := t.getOrgInfoById(stub, orgId)

	for i, order := range orders {
		//先判断一手，疫苗数量和填写的数量是否匹配
		fmt.Println("疫苗数量：" + strconv.Itoa(len(order.VaccineIds)))
		fmt.Println("疫苗数量：" + order.VaccNum)
		if strconv.Itoa(len(order.VaccineIds)) != order.VaccNum {
			return shim.Error("vaccnumNoMatch:" + strconv.Itoa(i+1))
		}

		stamp := time.Now().Unix()
		orderId := strconv.FormatInt(stamp, 10) + strconv.Itoa(i)
		order.OrderId = orgId + orderId
		buyer, _ := t.getOrgInfoById(stub, order.BuyerId)

		//order.OrderId = orgId + strconv.Itoa(i)
		order.BuyerId = buyer.Id
		order.BuyerName = buyer.Name
		order.BuyerSite = buyer.Address
		order.SellerId = orgId
		order.SellerName = seller.Name
		order.SellerSite = seller.Address
		order.Producer = seller.Name
		order.Owner = orgId
		order.Date = order.Date
		order.State = 0
		orderStr, _ := json.Marshal(order)
		stub.PutState(order.OrderId, []byte(orderStr))
		fmt.Println("A订单: " + string(orderStr))

		//创建联合主键M
		indexKey, _ := stub.CreateCompositeKey(combinedConstruction_M, []string{order.BuyerId, order.VaccineIds[0]})
		stub.PutState(indexKey, orderStr)

		for _, v := range order.VaccineIds {
			//判断是否疫苗是否合格
			if vacc, err := t.getVaccineById(stub, "Vaccine"+v); err != nil {
				return shim.Error("noVaccine:" + v)
			} else {
				if vacc.CheckState != "1" {
					return shim.Error("discheck:" + v)
				}
			}

			//减库存
			indexKey2, _ := stub.CreateCompositeKey(combinedConstruction_Inventory, []string{orgId, v})
			//判断是否有库存
			resultIterator, _ := stub.GetStateByPartialCompositeKey(combinedConstruction_Inventory, []string{orgId, v})
			defer resultIterator.Close()
			if !resultIterator.HasNext() {
				return shim.Error("noInventory:" + v)
			} else {
				stub.DelState(indexKey2)
				fmt.Println("减库存: ", v)
			}
		}

		//模拟发货并记录溯源信息
		for _, value := range order.VaccineIds {
			var trace Trace
			trace.Id = value + trace_suf
			trace.Name = order.VaccineName
			trace.Date = order.Date
			trace.Site = order.SellerSite
			trace.Owner = orgId
			trace.Action = "批发销售"
			trace.Result = "合格一致"
			traceStr, _ := json.Marshal(trace)
			stub.PutState(trace.Id, []byte(traceStr))
			fmt.Println("追溯: " + string(traceStr))
		}
	}
	return shim.Success([]byte("success"))
}

//查库存
func (t *hwChaincode) queryInventory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	orgId := args[0]
	resultIterator, _ := stub.GetStateByPartialCompositeKey(combinedConstruction_Inventory, []string{orgId})
	defer resultIterator.Close()
	var vaccines []Vaccine
	for resultIterator.HasNext() {
		item, _ := resultIterator.Next()
		var vacc Vaccine
		json.Unmarshal([]byte(item.Value), &vacc)
		vaccines = append(vaccines, vacc)
	}
	jsonStr, _ := json.Marshal(vaccines)
	fmt.Println("库存: " + string(jsonStr))
	return shim.Success(jsonStr)
}

//销售商添加订单addOrderB.html-采购验收
func (t *hwChaincode) addOrderB(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	orgId := args[0]
	orderStrs := args[1]
	var orders []OrderForm
	if err := json.Unmarshal([]byte(orderStrs), &orders); err != nil {
		return shim.Error("json反序列化失败")
	}
	buyer, _ := t.getOrgInfoById(stub, orgId)

	for i, order := range orders {
		//先判断一手，疫苗数量和填写的数量是否匹配
		fmt.Println("疫苗数量：" + strconv.Itoa(len(order.VaccineIds)))
		fmt.Println("疫苗数量：" + order.VaccNum)
		if strconv.Itoa(len(order.VaccineIds)) != order.VaccNum {
			return shim.Error("vaccnumNoMatch:" + strconv.Itoa(i+1))
		}

		stamp := time.Now().Unix()
		orderId := strconv.FormatInt(stamp, 10) + strconv.Itoa(i)
		order.OrderId = orgId + orderId
		seller, _ := t.getOrgInfoById(stub, order.SellerId)

		//order.OrderId = orgId + strconv.Itoa(i)
		order.BuyerId = buyer.Id
		order.BuyerName = buyer.Name
		order.BuyerSite = buyer.Address
		order.SellerId = seller.Id
		order.SellerName = seller.Name
		order.SellerSite = seller.Address
		order.Producer = seller.Name
		order.Owner = orgId
		order.Date = order.Date
		order.State = 0
		orderStr, _ := json.Marshal(order)
		stub.PutState(order.OrderId, []byte(orderStr))
		fmt.Println("B订单: " + string(orderStr))

		//模拟采购验收并记录溯源信息+加库存
		for _, value := range order.VaccineIds {

			vacc, err := t.getVaccineById(stub, "Vaccine"+value)
			//先验证疫苗是否检验合格
			if err != nil {
				return shim.Error("noVaccine:" + value)
			} else {
				if vacc.CheckState != "1" {
					return shim.Error("discheck:" + value)
				}
			}
			//加库存
			vaccJson, _ := json.Marshal(vacc)
			indexKey, _ := stub.CreateCompositeKey(combinedConstruction_Inventory, []string{orgId, value})
			stub.PutState(indexKey, []byte(vaccJson))

			var trace Trace
			trace.Id = value + trace_suf
			trace.Name = order.VaccineName
			trace.Date = order.Date
			trace.Site = order.BuyerSite
			trace.Owner = orgId
			trace.Action = "采购验收"
			trace.Result = "合格一致"
			traceStr, _ := json.Marshal(trace)
			stub.PutState(trace.Id, []byte(traceStr))
			fmt.Println("追溯: " + string(traceStr))
		}
	}
	return shim.Success([]byte("success"))
}

//销售商修改订单updateOrderB.html
func (t *hwChaincode) updateOrderB(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	orderId := args[0]     //订单Id
	VaccNum := args[1]     //订单疫苗数量
	VaccineName := args[2] //疫苗名称
	SellerId := args[3]    //生产商Id

	seller, _ := t.getOrgInfoById(stub, SellerId)
	order, err := t.getOrderFormById(stub, orderId)
	if err != nil {
		shim.Error("not this Order")
	}
	order.VaccNum = VaccNum
	order.VaccineName = VaccineName
	order.SellerId = SellerId
	order.SellerName = seller.Name
	order.SellerSite = seller.Address
	order.Producer = SellerId
	order.Owner = SellerId
	order.State = 0

	orderStr, _ := json.Marshal(order)
	stub.PutState(order.OrderId, []byte(orderStr))
	fmt.Println("B订单: " + string(orderStr))

	return shim.Success([]byte("success"))
}

//B公司申请背书
func (t *hwChaincode) SendEndorse(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	orderId := args[0]
	orderForm, err := t.getOrderFormById(stub, orderId)
	if err != nil {
		shim.Error("not this Order")
	}
	//更新订单背书状态为背书中
	orderForm.State = 1
	orderStr, _ := json.Marshal(orderForm)
	err = stub.PutState(orderId, orderStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("申请背书: " + string(orderStr))
	//创建联合主键N，可以让卖家看到所有需要背书的订单
	indexKey, err := stub.CreateCompositeKey(combinedConstruction_N, []string{orderForm.SellerId, orderForm.VaccineIds[0]})

	err = stub.PutState(indexKey, orderStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("success"))
}

//确认背书
func (t *hwChaincode) comfirmEndorse(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	orderId := args[0]
	endoserState := args[1]
	orderB, err := t.getOrderFormById(stub, orderId)
	if err != nil {
		shim.Error("not this Order")
	}

	resultIterator, _ := stub.GetStateByPartialCompositeKey(combinedConstruction_M, []string{orderB.BuyerId, orderB.VaccineIds[0]})
	defer resultIterator.Close()
	var orderA OrderForm
	var indexKey string
	//flag := false
	for resultIterator.HasNext() {
		//flag = true
		item, err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		indexKey = item.Key
		if err := json.Unmarshal(item.Value, &orderA); err != nil {
			return shim.Error("json Unmarshal faild")
		}
		fmt.Println("A订单信息: " + string(item.Value))
	}

	abc, _ := json.Marshal(orderB)
	fmt.Println("B订单信息: " + string(abc))

	//背书比对两份订单的VaccNum,相同背书成功
	//if orderB.VaccNum == orderA.VaccNum && orderA.VaccineName == orderB.VaccineName && flag {
	if endoserState == "0" {
		orderA.State = 2
		orderB.State = 2
		//更新疫苗状态及所有人信息，模拟收货并记录溯源信息
		for _, value := range orderA.VaccineIds {
			vacc, err := t.getVaccineById(stub, "Vaccine"+value)
			if err != nil {
				return shim.Error("not this vaccine")
			}
			vacc.State = 1
			vacc.Owner = orderB.BuyerId
			vaccJson, _ := json.Marshal(vacc)
			stub.PutState(vacc.Id, []byte(vaccJson))
			fmt.Println("更新后疫苗" + vacc.Id + "信息: " + string(vaccJson))

		}
		//更新A订单
		orderAJson, _ := json.Marshal(orderA)
		stub.PutState(orderA.OrderId, []byte(orderAJson))
		fmt.Println("更新后A订单信息: " + string(orderAJson))
		//更新B订单
		orderBJson, _ := json.Marshal(orderB)
		stub.PutState(orderB.OrderId, []byte(orderBJson))
		fmt.Println("更新后B订单信息: " + string(orderBJson))

		//删除联合主键M和N
		err = stub.DelState(indexKey)
		fmt.Println("delete Combine M：" + indexKey)
		if err != nil {
			return shim.Error(err.Error())
		}

		indexKeyN, err := stub.CreateCompositeKey(combinedConstruction_N, []string{orderB.SellerId, orderB.VaccineIds[0]})
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.DelState(indexKeyN)
		fmt.Println("delete Combine N：" + indexKeyN)
		return shim.Success([]byte("success"))
	} else {
		orderB.State = 3
		//更新B订单
		orderBJson, _ := json.Marshal(orderB)
		stub.PutState(orderB.OrderId, []byte(orderBJson))
		fmt.Println("更新后B订单信息: " + string(orderBJson))

		//删除联合主键N
		indexKeyN, err := stub.CreateCompositeKey(combinedConstruction_N, []string{orderB.SellerId, orderB.VaccineIds[0]})
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.DelState(indexKeyN)
		fmt.Println("背书失败：delete Combine N：" + indexKeyN)
		return shim.Success([]byte("Endorse faild:VaccNum faild Or VaccineName faild"))
	}

}

//销售商录入销售信息addSellInfo.html
func (t *hwChaincode) addSellInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//二个参数：卖家Id，json
	orgId := args[0]
	jsonStr := args[1]
	org, _ := t.getOrgInfoById(stub, orgId)
	var sellinfos []SellInfo
	json.Unmarshal([]byte(jsonStr), &sellinfos)

	for i, value := range sellinfos {

		sellInfo := value
		if vaccc, err := t.getVaccineById(stub, "Vaccine"+sellInfo.VaccId); err != nil {
			return shim.Error("noVaccine:" + sellInfo.VaccId)
		} else {
			if vaccc.Owner != orgId {
				return shim.Error("disendorse:" + sellInfo.VaccId)
			}
		}

		var vacc Vaccine
		var indexKey string
		resultIterator, _ := stub.GetStateByPartialCompositeKey(combinedConstruction_Inventory, []string{orgId, sellInfo.VaccId})
		defer resultIterator.Close()

		if !resultIterator.HasNext() {
			return shim.Error("noInventory:" + sellInfo.VaccId)
		}
		for resultIterator.HasNext() {
			item, _ := resultIterator.Next()
			indexKey = item.Key
			json.Unmarshal([]byte(item.Value), &vacc)
		}

		oldcount, _ := strconv.Atoi(vacc.Count)
		newcount, _ := strconv.Atoi(sellInfo.VaccNum)

		if (oldcount - newcount) >= 0 {

			//更新库存
			vacc.Count = strconv.Itoa(oldcount - newcount)
			vaccjsonStr, _ := json.Marshal(vacc)
			stub.PutState(indexKey, vaccjsonStr)

			stamp := time.Now().Unix()
			sellInfoId := strconv.FormatInt(stamp, 10) + strconv.Itoa(i)
			//添加销售订单
			sellInfo.Id = "sell-" + sellInfoId
			sellInfo.SellerId = orgId
			sellInfo.SellerName = org.Name
			sellInfo.SellerSite = org.Address
			sellInfo.Date = sellInfo.Date
			sellInfoJson, _ := json.Marshal(sellInfo)
			stub.PutState(sellInfo.Id, sellInfoJson)

			//添加追溯信息
			var trace Trace
			trace.Id = sellInfo.VaccId + trace_suf
			trace.Name = vacc.Name
			trace.Date = sellInfo.Date
			trace.Site = sellInfo.Site
			trace.Owner = sellInfo.BuyerName
			trace.Action = "终端售出"
			trace.Result = "合格一致"
			traceStr, _ := json.Marshal(trace)
			stub.PutState(trace.Id, []byte(traceStr))
			fmt.Println("追溯: " + string(traceStr))
		} else {
			return shim.Error("InventoryShortages:" + sellInfo.VaccId)
		}
	}

	return shim.Success([]byte("success"))
}

//生产商查询所有需要自己背书的订单
func (t *hwChaincode) queryUnEndorseOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	orgId := args[0]
	resultIterator, err := stub.GetStateByPartialCompositeKey(combinedConstruction_N, []string{orgId})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultIterator.Close()

	var orders []OrderForm
	for resultIterator.HasNext() {
		item, err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var order OrderForm
		if err := json.Unmarshal(item.Value, &order); err != nil {
			return shim.Error("json Unmarshal faild")
		}
		orders = append(orders, order)
	}

	ordersJson, _ := json.Marshal(orders)
	fmt.Println("待背书列表：" + string(ordersJson))
	return shim.Success(ordersJson)
}

//查询订单，获取详情供orderFormA.html,orderFormB.html两页面使用
func (t *hwChaincode) queryOrderForm(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	orderId := args[0]
	orderForm, err := t.getOrderFormById(stub, orderId)
	if err != nil {
		shim.Error("not this OrderForm")
	}
	orderFormStr, _ := json.Marshal(orderForm)
	fmt.Println("订单" + orderId + ": " + string(orderFormStr))
	return shim.Success(orderFormStr)
}

//查询订单列表根据参数取A公司或者B公司
func (s *hwChaincode) queryAllVaccineList(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	orgId := args[0]
	startKey := "Vaccine0"
	endKey := "Vaccine9999999999999"
	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var vaccines []Vaccine
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var vacc Vaccine
		json.Unmarshal([]byte(queryResponse.Value), &vacc)
		vaccines = append(vaccines, vacc)
	}
	orderFormsJson, _ := json.Marshal(vaccines)
	fmt.Println(orgId + "公司疫苗列表:" + string(orderFormsJson))
	return shim.Success(orderFormsJson)
}

//查询订单列表根据参数取A公司或者B公司
func (s *hwChaincode) queryAllSellInfoList(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	orgId := args[0]
	startKey := "sell-" + "0"
	endKey := "sell-" + "9999999999999"
	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var sellInfos []SellInfo
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var sellInfo SellInfo
		json.Unmarshal([]byte(queryResponse.Value), &sellInfo)
		sellInfos = append(sellInfos, sellInfo)
	}
	orderFormsJson, _ := json.Marshal(sellInfos)
	fmt.Println(orgId + "公司销售订单列表:" + string(orderFormsJson))
	return shim.Success(orderFormsJson)
}

//查询订单列表根据参数取A公司或者B公司
func (s *hwChaincode) queryAllOrderList(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	orgId := args[0]
	startKey := orgId + "0"
	endKey := orgId + "99999999999"
	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var orderForms []OrderForm
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var orderForm OrderForm
		json.Unmarshal([]byte(queryResponse.Value), &orderForm)
		if orderForm.OrderId != "" {
			orderForms = append(orderForms, orderForm)
		}
	}
	orderFormsJson, _ := json.Marshal(orderForms)
	fmt.Println(orgId + "公司订单列表:" + string(orderFormsJson))
	return shim.Success(orderFormsJson)
}

//查询订单列表根据参数取A公司或者B公司
func (s *hwChaincode) queryTraceHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	vaccId := args[0]
	vacc, _ := s.getVaccineById(stub, "Vaccine"+vaccId)

	vaccStr, _ := json.Marshal(vacc)

	org, _ := s.getOrgInfoById(stub, vacc.Producer)
	orgStr, _ := json.Marshal(org)

	iter, err := stub.GetHistoryForKey(vaccId + trace_suf)
	if err != nil {
		shim.Error("not Trace History")
	}
	defer iter.Close()
	var traces []Trace
	for iter.HasNext() {
		res, err := iter.Next()
		if err != nil {
			shim.Error("not Trace History")
		}
		var trace Trace
		json.Unmarshal([]byte(res.Value), &trace)
		traces = append(traces, trace)
	}

	tracesStr, _ := json.Marshal(traces)

	jsonStr := "{\"vaccine\":" + string(vaccStr) + ",\"org\":" + string(orgStr) + ",\"trace\":" + string(tracesStr) + "}"

	fmt.Println("公司订单列表:" + jsonStr)
	return shim.Success([]byte(jsonStr))

}

func (t *hwChaincode) submitEvaluate(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("参数出错")
	}
	evalStr := args[0]
	var eval Evaluate
	if err := json.Unmarshal([]byte(evalStr), &eval); err != nil {
		return shim.Error("json deserilized faild!")
	}

	stamp := time.Now().Unix()
	eval.Id = strconv.FormatInt(stamp, 10)
	evalJsonStr, _ := json.Marshal(eval)
	stub.PutState(eval.Id, evalJsonStr)
	fmt.Println("评价详情:" + string(evalJsonStr))

	//添加追溯信息
	if strings.Index(string(evalJsonStr), "不一致") == -1 {
		var trace Trace
		trace.Id = eval.VaccId + trace_suf
		trace.Date = time.Now().Format(dateFomat)
		trace.Site = "追溯链条已合成"
		trace.Action = "消费评价"
		trace.Result = "合格一致"
		traceStr, _ := json.Marshal(trace)
		stub.PutState(trace.Id, []byte(traceStr))
		fmt.Println("追溯: " + string(traceStr))
	} else {
		var trace Trace
		trace.Id = eval.VaccId + trace_suf
		trace.Date = time.Now().Format(dateFomat)
		trace.Site = "追溯链条未合成"
		trace.Action = "消费评价"
		if eval.EvaBaseInfo == "不一致" {
			trace.Result += "基本信息不一致;"
		}
		if eval.EvaSellDate == "不一致" {
			trace.Result += "购买时间不一致;"
		}
		if eval.EvaSellSite == "不一致" {
			trace.Result += "购买地点不一致;"
		}
		if eval.EvaSellPlace == "不一致" {
			trace.Result += "购买场所不一致;"
		}
		if eval.EvaSellState == "不一致" {
			trace.Result += "购买状态不一致;"
		}
		if eval.Complain != "" {
			trace.Result += eval.Complain
		}
		traceStr, _ := json.Marshal(trace)
		stub.PutState(trace.Id, []byte(traceStr))
		fmt.Println("追溯: " + string(traceStr))
	}

	return shim.Success([]byte("success"))
}

/*******通用方法区*******/
func (t *hwChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

//根据id获取企业
func (t *hwChaincode) getOrgInfoById(stub shim.ChaincodeStubInterface, id string) (OrgInfo, error) {
	dvalue, err := stub.GetState(id)
	var d OrgInfo
	err = json.Unmarshal([]byte(dvalue), &d)
	return d, err
}

//根据Id获取订单
func (t *hwChaincode) getOrderFormById(stub shim.ChaincodeStubInterface, id string) (OrderForm, error) {
	dvalue, err := stub.GetState(id)
	var d OrderForm
	err = json.Unmarshal([]byte(dvalue), &d)
	return d, err
}

//根据Id获取疫苗
func (t *hwChaincode) getVaccineById(stub shim.ChaincodeStubInterface, id string) (Vaccine, error) {
	dvalue, err := stub.GetState(id)
	var d Vaccine
	err = json.Unmarshal([]byte(dvalue), &d)
	return d, err
}

func isExisted(stub shim.ChaincodeStubInterface, key string) bool {
	val, err := stub.GetState(key)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	if len(val) == 0 {
		return false
	}
	return true
}

//生成32位md5字串
func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func uniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getMd5String(base64.URLEncoding.EncodeToString(b))
}

func main() {
	err := shim.Start(new(hwChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
