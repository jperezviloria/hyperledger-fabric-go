package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//===========================./fabric-samples/chaincode-docker-devmode/docker-compose-simple.yaml===========
//si queremos levantar el la network en devmode corremos el comando siguiente
//docker-compose -f docker-compose-simple.yaml up
//OJO es posible que algunos contenedores esten inactivos en background, para eliminarlos usar el comando
//docker rm $(docker ps -a -f status=exited -f status=created -q)
//========================== CHAINCODE dentro de el directorio ./fabric-samples/ correr el comando ===================
//sudo chmod -R 777 .
//docker exec -it chaincode sh
//este nos llevara a una shell dentro del directorio ./chaincode, nos dirigiremos a archivo donde se ubica nuestro archivo chaincode
//ejecutamos el comando -> go mod init example.com/usuario, de esta forma generara un archivo go.mod
//luego ejecutamos el comando -> go build, se generara un archivo go.sum y un ejecutable
//OJO si no se genera dentro del codigo una funcion main dentro de un package main no se generara el ejecutable
//luego correr para iniciar el programa con el ejecutable
//CORE_CHAINCODE_ID_NAME=us:0 CORE_PEER_TLS_ENABLED=false ./usuario -peer.address peer:7052
//========================== CLI dentro de el directorio ./fabric-samples/ correr el comando ===================
//correr -> sudo docker exec -it cli bash
//-> cd chaincode
//-> peer chaincode install -p huenei-pruebas/go -n us -v 0
//-> peer chaincode instantiate -n us -v 0 -c '{"Args":[]}' -C myc
//-> peer chaincode invoke -n us -c '{"Args":["CreateEmployer","100","Julio","1000"]}' -C myc
//-> peer chaincode invoke -n us -c '{"Args":["CreateEmployer","200","Milton","2000"]}' -C myc
//-> peer chaincode invoke -n us -c '{"Args":["CreateEmployer","300","Nicolas","3000"]}' -C myc
//-> peer chaincode invoke -n us -c '{"Args":["GetAllEmployers"]}' -C myc
//-> peer chaincode invoke -n us -c '{"Args":["GetEmployerById","100"]}' -C myc
//===========================TEST NETWORK=====================================================================
//--- Test-network
//cd fabric-samples/test-network
//
//sudo chmod -R 777 .
//
//sudo ./network.sh down
//sudo ./network.sh up -ca -s couchdb
//
//sudo ./network.sh createChannel -c testchannel
//
//Estando como root (sudo -i) (de otra forma no funcionó)
//
//export GOPATH=/home/protobot/go
//export GOROOT=/home/protobot/sdk/go1.15.3
//export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
//
//cd home/msesarego/fabric-samples/test-network
//
//source ./chaincodeLifecycleHueneiPuntos.sh && start  (del lado izquierdo es el nombre del archivo y de derecho es el nombre de la funcion shell a ejecutar)
//source ./chaincodeLifecycleHueneiPuntos.sh && chaincodeInitLedger
//source ./chaincodeLifecycleHueneiPuntos.sh && chaincodeAgregarUsuario
//source ./chaincodeLifecycleHueneiPuntos.sh && chaincodeQueryAll
//source ./chaincodeLifecycleHueneiPuntos.sh && chaincodeTransferirPuntosPorBuenDesempenio

type SmartContractDefinition struct {
	contractapi.Contract
}

type Employer struct {
	IdEmployer     string `json:"idEmployer"`
	NameEmployer   string `json:"nameEmployer"`
	WalletEmployer int    `json:"walletEmployer"`
}

func (smartContract *SmartContractDefinition) GetAllEmployers(ctx contractapi.TransactionContextInterface) ([]*Employer, error) {

	//here we declaring the iteration with all ledger, startKey value is inclusive but endKey value is exclusive
	resultIterator, err := ctx.GetStub().GetStateByRange("", "")
	//every declarations and actions have been a condition if declaration is nil(null)
	if err != nil {
		return nil, err
	}
	//defer is a Go keyword that close the process before finish
	defer resultIterator.Close()

	//this is the real Employers information
	var employers []*Employer
	//using a a for to iterating with each employer
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}
		var particularEmployer Employer
		err = json.Unmarshal(queryResponse.Value, &particularEmployer)
		if err != nil {
			return nil, err
		}
		employers = append(employers, &particularEmployer)
	}
	fmt.Print("The Employers was found successfully")
	return employers, nil
}

func (smartContract *SmartContractDefinition) GetEmployerById(ctx contractapi.TransactionContextInterface, idEmployer string) (*Employer, error) {

	employerJSON, err := ctx.GetStub().GetState(idEmployer)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	//TO DO if employer dont exist the code need response with a message
	/*if employerJSON != nil {
		return nil, fmt.Errorf("the employerJSON %s does not exist", idEmployer)
	}*/

	var particularEmployer Employer
	err = json.Unmarshal(employerJSON, &particularEmployer)
	if err != nil {
		return nil, err
	}
	fmt.Print("The Employer was found successfully")
	return &particularEmployer, nil

}

func (smartContract *SmartContractDefinition) CreateEmployer(
	ctx contractapi.TransactionContextInterface,
	idEmployer string,
	nameEmployer string,
	walletEmployer int) error {

	//here we are creating entity with out validate
	employer := Employer{
		IdEmployer:     idEmployer,
		NameEmployer:   nameEmployer,
		WalletEmployer: walletEmployer,
	}
	employerJSON, err := json.Marshal(employer)
	if err != nil {
		return err
	}
	fmt.Print("The Employer was created successfully")
	return ctx.GetStub().PutState(idEmployer, employerJSON)
}

func (smartContract *SmartContractDefinition) CreateEmployerWithValidation(
	ctx contractapi.TransactionContextInterface,
	idEmployer string,
	nameEmployer string,
	walletEmployer int) error {

	//here we`re start to validate that entity was created
	validateEmployer, err := smartContract.EmployerExist(ctx, idEmployer)
	if err != nil {
		return err
	}
	if validateEmployer {
		return fmt.Errorf("the employer $s already exist", idEmployer)
	}

	//here we`re start to create entity
	employer := Employer{
		IdEmployer:     idEmployer,
		NameEmployer:   nameEmployer,
		WalletEmployer: walletEmployer,
	}
	employerJSON, err := json.Marshal(employer)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(idEmployer, employerJSON)
}

func (smartContract *SmartContractDefinition) EmployerExist(ctx contractapi.TransactionContextInterface, idEmployer string) (bool, error) {
	employerJSON, err := ctx.GetStub().GetState(idEmployer)
	if err != nil {
		return false, fmt.Errorf("failed to read frm world state: %v", err)
	}
	return employerJSON != nil, nil
}

func (smartContract *SmartContractDefinition) TransferAssetWallet(
	ctx contractapi.TransactionContextInterface,
	idEmployerFrom string,
	idEmployerTo string,
	valueToSend int) error {

	validateEmployerFrom, err := smartContract.GetEmployerById(ctx, idEmployerFrom)
	if err != nil {
		return err
	}

	validateEmployerTo, err := smartContract.GetEmployerById(ctx, idEmployerTo)
	if err != nil {
		return err
	}
	validateEmployerFrom.WalletEmployer = validateEmployerFrom.WalletEmployer - valueToSend
	validateEmployerTo.WalletEmployer = validateEmployerTo.WalletEmployer + valueToSend

	employerFromJSON, err := json.Marshal(validateEmployerFrom)
	if err != nil {
		return nil
	}
	employerToJSON, err := json.Marshal(validateEmployerTo)
	if err != nil {
		return nil
	}
	ctx.GetStub().PutState(idEmployerFrom, employerFromJSON)
	ctx.GetStub().PutState(idEmployerTo, employerToJSON)

	return nil
}

func main() {

	smartContractDefinition := new(SmartContractDefinition)

	chaincode, err := contractapi.NewChaincode(smartContractDefinition)
	if err != nil {
		panic(err.Error())
	}
	if err := chaincode.Start(); err != nil {
		panic(err.Error())
	}
}
