package helpers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/pkg/errors"
	"github.com/savaki/jq"
)

// ITemplateParser describes config variables parser.
type ITemplateParser interface {
	Compile(expression string) (ITemplateExpression, error)
}

// ITemplateExpression descries single pre-compiled expression.
type ITemplateExpression interface {
	Parse(payload string) (interface{}, error)
	Format(params map[string]interface{}) (interface{}, error)
}

// Parser implementation.
type parser struct {
	functions map[string]govaluate.ExpressionFunction
}

// Parser expression
type parserExpression struct {
	expression *govaluate.EvaluableExpression
}

// NewParser constructs a new template parser.
func NewParser() ITemplateParser {
	p := &parser{
		functions: map[string]govaluate.ExpressionFunction{
			"jq":  jqParse,
			"num": float64Convert,
			"str": strConvert,
			"fmt": format,
		},
	}

	return p
}

// Compile tries to pre-compile expression.
func (p *parser) Compile(expression string) (ITemplateExpression, error) {
	exp, err := govaluate.NewEvaluableExpressionWithFunctions(expression, p.functions)
	if err != nil {
		return nil, errors.Wrap(err, "govaluate compilation failed")
	}

	return &parserExpression{
		expression: exp,
	}, nil
}

// Parse tries to parse received payload into internal param.
func (p *parserExpression) Parse(payload string) (interface{}, error) {
	params := map[string]interface{}{"payload": payload}
	return p.expression.Evaluate(params)
}

// Format tries to convert internal param into output payload.
func (p *parserExpression) Format(params map[string]interface{}) (interface{}, error) {
	if params == nil {
		params = make(map[string]interface{})
	}

	data, err := p.expression.Evaluate(params)
	if err != nil {
		return nil, errors.Wrap(err, "govaluate evaluation failed")
	}

	return data, nil
}

// If only one param is supplied, returns un-marshaled json object.
// If two params are supplied, regular JQ syntax is used.
func jqParse(arguments ...interface{}) (interface{}, error) {
	if 0 == len(arguments) {
		return nil, &ErrArgumentsMismatch{Count: len(arguments)}
	}

	arg1, ok := arguments[0].(string)
	if !ok {
		return nil, &ErrWrongArgument{Message: "first argument is not a string"}
	}

	if 1 == len(arguments) {
		data := make(map[string]interface{})
		err := json.Unmarshal([]byte(arg1), &data)
		if err != nil {
			return nil, errors.Wrap(err, "json un-marshal failed")
		}

		return data, nil
	}

	if 2 == len(arguments) {
		arg2, ok := arguments[1].(string)
		if !ok {
			return nil, &ErrWrongArgument{Message: "second argument is not a string"}
		}

		op, err := jq.Parse(arg2)
		if err != nil {
			return nil, &ErrJqSyntax{Message: "failed to parse"}
		}

		val, err := op.Apply([]byte(arg1))
		if err != nil {
			return nil, &ErrJqSyntax{Message: "failed to apply"}
		}

		return strings.Trim(string(val), "\""), nil
	}

	return nil, &ErrArgumentsMismatch{Count: len(arguments)}
}

// Converts input param into int.
func float64Convert(arguments ...interface{}) (interface{}, error) {
	if 1 != len(arguments) {
		return nil, &ErrArgumentsMismatch{Count: len(arguments)}
	}

	if reflect.TypeOf(arguments[0]).Kind() == reflect.String {
		v, err := strconv.ParseFloat(arguments[0].(string), 64)
		if err != nil {
			return nil, errors.Wrap(err, "string conversion failed")
		}
		return v, nil
	}

	a, ok := arguments[0].(float64)
	if !ok {
		return nil, &ErrWrongArgument{Message: "not compatible with int type"}
	}

	return a, nil
}

// Converts input param into string.
func strConvert(arguments ...interface{}) (interface{}, error) {
	if 1 != len(arguments) {
		return nil, &ErrArgumentsMismatch{Count: len(arguments)}
	}

	a, ok := arguments[0].(string)
	if !ok {
		return fmt.Sprintf("%v", arguments[0]), nil
	}

	return a, nil
}

// Uses fmt.Sprintf.
func format(arguments ...interface{}) (interface{}, error) {
	if 0 == len(arguments) {
		return nil, &ErrArgumentsMismatch{Count: len(arguments)}
	}

	if 1 == len(arguments) {
		return strConvert(arguments[0])
	}

	a, err := strConvert(arguments[0])
	if err != nil {
		return nil, &ErrWrongArgument{Message: "not compatible with string type"}
	}

	return fmt.Sprintf(a.(string), arguments[1:]...), nil
}
