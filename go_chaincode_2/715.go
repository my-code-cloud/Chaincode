package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dovetail-lab/fabric-chaincode/common"
	trigger "github.com/dovetail-lab/fabric-chaincode/trigger/transaction"
	shim "github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/project-flogo/core/app"
	_ "github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
)

const (
	fabricTrigger = "#transaction"
)

// Contract implements chaincode interface for invoking Flogo flows
type Contract struct {
}

var logger = log.ChildLogger(log.RootLogger(), "fabric-cli-shim")

func init() {
	//  get log level from env FLOGO_LOG_LEVEL or CORE_CHAINCODE_LOGGING_LEVEL
	logLevel := "DEBUG"
	if l, ok := os.LookupEnv("FLOGO_LOG_LEVEL"); ok {
		logLevel = strings.ToUpper(l)
	} else if l, ok := os.LookupEnv("CORE_CHAINCODE_LOGGING_LEVEL"); ok {
		logLevel = strings.ToUpper(l)
	}
	switch logLevel {
	case "FATAL", "PANIC", "ERROR":
		log.SetLogLevel(log.RootLogger(), log.ErrorLevel)
	case "WARN", "WARNING":
		log.SetLogLevel(log.RootLogger(), log.WarnLevel)
	case "INFO":
		log.SetLogLevel(log.RootLogger(), log.InfoLevel)
	case "DEBUG", "TRACE":
		log.SetLogLevel(log.RootLogger(), log.DebugLevel)
	default:
		log.SetLogLevel(log.RootLogger(), log.DefaultLogLevel)
	}
}

// Init is called during chaincode instantiation to initialize any data,
// and also calls this function to reset or to migrate data.
func (t *Contract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode.
func (t *Contract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()
	logger.Debugf("invoke transaction fn=%s, args=%+v", fn, args)

	trig, ok := trigger.GetTrigger(fn)
	if !ok {
		return shim.Error(fmt.Sprintf("function %s is not implemented", fn))
	}
	status, result, err := trig.Invoke(stub, fn, args)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to execute transaction: %s, error: %+v", fn, err))
	} else if status == shim.OK {
		return shim.Success([]byte(result))
	} else {
		return pb.Response{
			Status:  int32(status),
			Payload: []byte(result),
		}
	}
}

var (
	cfgJson       string
	cfgEngine     string
	cfgCompressed bool
)

// main function starts up the chaincode in the container during instantiate
func main() {

	os.Setenv("FLOGO_RUNNER_TYPE", "DIRECT")
	os.Setenv("FLOGO_ENGINE_STOP_ON_ERROR", "false")

	// necessary to access schema of complex object attributes from activity context
	schema.Enable()
	schema.DisableValidation()

	cfg, err := engine.LoadAppConfig(cfgJson, cfgCompressed)
	if err != nil {
		logger.Errorf("Failed to load flogo config: %s", err.Error())
		os.Exit(1)
	}

	// set mapping to pass fabric stub to activities in the flow
	// this is a workaround until flogo-lib can accept pass-through flow attributes in
	// handler.Handle(context.Background(), triggerData) that bypasses the mapper.
	// see issue: https://github.com/TIBCOSoftware/flogo-lib/issues/267
	inputAssignMap(cfg, fabricTrigger, common.FabricStub)

	_, err = engine.New(cfg, engine.ConfigOption(cfgEngine, cfgCompressed))
	if err != nil {
		logger.Errorf("Failed to create flogo engine instance: %+v", err)
		os.Exit(1)
	}

	if err := shim.Start(new(Contract)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

// inputAssignMap sets additional input mapping for a specified trigger ref type;
// this is to ensure the mapping of a trigger output property to avoid user error.
func inputAssignMap(ac *app.Config, triggerRef, name string) {
	// add the name to all flow resources
	prop := map[string]interface{}{"name": name, "type": "any"}
	for _, rc := range ac.Resources {
		var jsonobj map[string]interface{}
		if err := json.Unmarshal(rc.Data, &jsonobj); err != nil {
			logger.Errorf("failed to parse resource data %s: %+v", rc.ID, err)
			continue
		}
		if metadata, ok := jsonobj["metadata"]; ok {
			metaMap := metadata.(map[string]interface{})
			if input, ok := metaMap["input"]; ok {
				inputArray := input.([]interface{})
				done := false
				for _, ip := range inputArray {
					ipMap := ip.(map[string]interface{})
					if ipMap["name"].(string) == name {
						done = true
						continue
					}
				}
				if !done {
					logger.Debugf("add new property %s to resource input of %s", name, rc.ID)
					metaMap["input"] = append(inputArray, prop)
					if jsonbytes, err := json.Marshal(jsonobj); err == nil {
						logger.Debugf("resource data is updated for %s: %s", rc.ID, string(jsonbytes))
						rc.Data = jsonbytes
					} else {
						logger.Debugf("failed to serialize resource %s: %+v", rc.ID, err)
					}
				}
			}
		}
	}
	// add input mapper
	for _, tc := range ac.Triggers {
		if tc.Ref == triggerRef {
			for _, hc := range tc.Handlers {
				for _, acc := range hc.Actions {
					if acc.Ref == "github.com/project-flogo/flow" {
						_, done := acc.Input[name]
						if !done {
							acc.Input[name] = "=$." + name
						}
					}
				}
			}
		}
	}
}
