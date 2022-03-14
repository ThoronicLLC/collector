package cel

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	log "github.com/sirupsen/logrus"
)

func ruleDetection(jsonString string, rules []string, logger *log.Entry) bool {
	// Loop through all the detection rules
	for _, rule := range rules {
		// Check if detected
		result, err := detectionLogic(jsonString, rule)

		// Log error if debug is set and a log entry was supplied
		if err != nil {
			if log.IsLevelEnabled(log.DebugLevel) {
				if logger != nil {
					logger.Errorf("issue with CEL rule: %s", err)
				}
			}
		}

		// Return true if a result was truthy
		if result {
			return true
		}
	}

	return false
}

func detectionLogic(jsonString, rule string) (bool, error) {
	// First build the CEL program.
	ds := cel.Declarations(
		decls.NewConst("event", decls.NewMapType(decls.String, decls.Dyn), nil),
	)

	// Create the environment
	env, err := cel.NewEnv(ds)
	if err != nil {
		return false, fmt.Errorf("issue creating cel environment: %v", err)
	}

	prs, iss := env.Parse(rule)
	if iss != nil && iss.Err() != nil {
		return false, fmt.Errorf("issue parsing rule: %v", err)
	}

	chk, iss := env.Check(prs)
	if iss != nil && iss.Err() != nil {
		return false, fmt.Errorf("issue checking rule: %v", err)
	}

	prg, err := env.Program(chk)
	if err != nil {
		return false, fmt.Errorf("issue creating program: %v", err)
	}

	var spb structpb.Struct
	if err := jsonpb.Unmarshal(bytes.NewBuffer([]byte(jsonString)), &spb); err != nil {
		return false, fmt.Errorf("jsonpb unmarshal failed: %v", err)
	}

	// Now, evaluate the program and check the output.
	val, _, err := prg.Eval(map[string]interface{}{"event": &spb})
	if err != nil {
		return false, fmt.Errorf("evaluation failed: %v", err)
	}

	// Handle boolean type conversion
	val.Type().TypeName()
	if val.Type().TypeName() == "bool" {
		if val.Value().(bool) == true {
			return true, nil
		}
	} else {
		return false, fmt.Errorf("value returned (type: %v); value(%v)", val.Type().TypeName(), val)
	}

	return false, nil
}

func validateRule(rule string) error {
	// First build the CEL program.
	ds := cel.Declarations(
		decls.NewConst("event", decls.NewMapType(decls.String, decls.Dyn), nil),
	)

	// Create the environment
	env, err := cel.NewEnv(ds)
	if err != nil {
		return fmt.Errorf("issue creating cel environment: %v", err)
	}

	_, iss := env.Parse(rule)
	if iss != nil && iss.Err() != nil {
		return fmt.Errorf("issue parsing rule: %v", err)
	}

	return nil
}
