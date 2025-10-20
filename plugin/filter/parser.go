package filter

import (
	"time"

	"github.com/pkg/errors"
	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func buildCondition(expr *exprv1.Expr, schema Schema) (Condition, error) {
	switch v := expr.ExprKind.(type) {
	case *exprv1.Expr_CallExpr:
		return buildCallCondition(v.CallExpr, schema)
	case *exprv1.Expr_ConstExpr:
		val, err := getConstValue(expr)
		if err != nil {
			return nil, err
		}
		switch v := val.(type) {
		case bool:
			return &ConstantCondition{Value: v}, nil
		case int64:
			return &ConstantCondition{Value: v != 0}, nil
		case float64:
			return &ConstantCondition{Value: v != 0}, nil
		default:
			return nil, errors.New("filter must evaluate to a boolean value")
		}
	case *exprv1.Expr_IdentExpr:
		name := v.IdentExpr.GetName()
		field, ok := schema.Field(name)
		if !ok {
			return nil, errors.Errorf("unknown identifier %q", name)
		}
		if field.Type != FieldTypeBool {
			return nil, errors.Errorf("identifier %q is not boolean", name)
		}
		return &FieldPredicateCondition{Field: name}, nil
	default:
		return nil, errors.New("unsupported top-level expression")
	}
}

func buildCallCondition(call *exprv1.Expr_Call, schema Schema) (Condition, error) {
	switch call.Function {
	case "_&&_":
		if len(call.Args) != 2 {
			return nil, errors.New("logical AND expects two arguments")
		}
		left, err := buildCondition(call.Args[0], schema)
		if err != nil {
			return nil, err
		}
		right, err := buildCondition(call.Args[1], schema)
		if err != nil {
			return nil, err
		}
		return &LogicalCondition{
			Operator: LogicalAnd,
			Left:     left,
			Right:    right,
		}, nil
	case "_||_":
		if len(call.Args) != 2 {
			return nil, errors.New("logical OR expects two arguments")
		}
		left, err := buildCondition(call.Args[0], schema)
		if err != nil {
			return nil, err
		}
		right, err := buildCondition(call.Args[1], schema)
		if err != nil {
			return nil, err
		}
		return &LogicalCondition{
			Operator: LogicalOr,
			Left:     left,
			Right:    right,
		}, nil
	case "!_":
		if len(call.Args) != 1 {
			return nil, errors.New("logical NOT expects one argument")
		}
		child, err := buildCondition(call.Args[0], schema)
		if err != nil {
			return nil, err
		}
		return &NotCondition{Expr: child}, nil
	case "_==_", "_!=_", "_<_", "_>_", "_<=_", "_>=_":
		return buildComparisonCondition(call, schema)
	case "@in":
		return buildInCondition(call, schema)
	case "contains":
		return buildContainsCondition(call, schema)
	default:
		val, ok, err := evaluateBool(call)
		if err != nil {
			return nil, err
		}
		if ok {
			return &ConstantCondition{Value: val}, nil
		}
		return nil, errors.Errorf("unsupported call expression %q", call.Function)
	}
}

func buildComparisonCondition(call *exprv1.Expr_Call, schema Schema) (Condition, error) {
	if len(call.Args) != 2 {
		return nil, errors.New("comparison expects two arguments")
	}
	op, err := toComparisonOperator(call.Function)
	if err != nil {
		return nil, err
	}

	left, err := buildValueExpr(call.Args[0], schema)
	if err != nil {
		return nil, err
	}
	right, err := buildValueExpr(call.Args[1], schema)
	if err != nil {
		return nil, err
	}

	// If the left side is a field, validate allowed operators.
	if field, ok := left.(*FieldRef); ok {
		def, exists := schema.Field(field.Name)
		if !exists {
			return nil, errors.Errorf("unknown identifier %q", field.Name)
		}
		if def.Kind == FieldKindVirtualAlias {
			def, exists = schema.ResolveAlias(field.Name)
			if !exists {
				return nil, errors.Errorf("invalid alias %q", field.Name)
			}
		}
		if def.AllowedComparisonOps != nil {
			if _, allowed := def.AllowedComparisonOps[op]; !allowed {
				return nil, errors.Errorf("operator %s not allowed for field %q", op, field.Name)
			}
		}
	}

	return &ComparisonCondition{
		Left:     left,
		Operator: op,
		Right:    right,
	}, nil
}

func buildInCondition(call *exprv1.Expr_Call, schema Schema) (Condition, error) {
	if len(call.Args) != 2 {
		return nil, errors.New("in operator expects two arguments")
	}

	// Handle identifier in list syntax.
	if identName, err := getIdentName(call.Args[0]); err == nil {
		if field, ok := schema.Field(identName); ok && field.Kind == FieldKindVirtualAlias {
			if _, aliasOk := schema.ResolveAlias(identName); !aliasOk {
				return nil, errors.Errorf("invalid alias %q", identName)
			}
		} else if !ok {
			return nil, errors.Errorf("unknown identifier %q", identName)
		}

		if listExpr := call.Args[1].GetListExpr(); listExpr != nil {
			values := make([]ValueExpr, 0, len(listExpr.Elements))
			for _, element := range listExpr.Elements {
				value, err := buildValueExpr(element, schema)
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
			return &InCondition{
				Left:   &FieldRef{Name: identName},
				Values: values,
			}, nil
		}
	}

	// Handle "value in identifier" syntax.
	if identName, err := getIdentName(call.Args[1]); err == nil {
		if _, ok := schema.Field(identName); !ok {
			return nil, errors.Errorf("unknown identifier %q", identName)
		}
		element, err := buildValueExpr(call.Args[0], schema)
		if err != nil {
			return nil, err
		}
		return &ElementInCondition{
			Element: element,
			Field:   identName,
		}, nil
	}

	return nil, errors.New("invalid use of in operator")
}

func buildContainsCondition(call *exprv1.Expr_Call, schema Schema) (Condition, error) {
	if call.Target == nil {
		return nil, errors.New("contains requires a target")
	}
	targetName, err := getIdentName(call.Target)
	if err != nil {
		return nil, err
	}

	field, ok := schema.Field(targetName)
	if !ok {
		return nil, errors.Errorf("unknown identifier %q", targetName)
	}
	if !field.SupportsContains {
		return nil, errors.Errorf("identifier %q does not support contains()", targetName)
	}
	if len(call.Args) != 1 {
		return nil, errors.New("contains expects exactly one argument")
	}
	value, err := getConstValue(call.Args[0])
	if err != nil {
		return nil, errors.Wrap(err, "contains only supports literal arguments")
	}
	str, ok := value.(string)
	if !ok {
		return nil, errors.New("contains argument must be a string")
	}
	return &ContainsCondition{
		Field: targetName,
		Value: str,
	}, nil
}

func buildValueExpr(expr *exprv1.Expr, schema Schema) (ValueExpr, error) {
	if identName, err := getIdentName(expr); err == nil {
		if _, ok := schema.Field(identName); !ok {
			return nil, errors.Errorf("unknown identifier %q", identName)
		}
		return &FieldRef{Name: identName}, nil
	}

	if literal, err := getConstValue(expr); err == nil {
		return &LiteralValue{Value: literal}, nil
	}

	if value, ok, err := evaluateNumeric(expr); err != nil {
		return nil, err
	} else if ok {
		return &LiteralValue{Value: value}, nil
	}

	if boolVal, ok, err := evaluateBoolExpr(expr); err != nil {
		return nil, err
	} else if ok {
		return &LiteralValue{Value: boolVal}, nil
	}

	if call := expr.GetCallExpr(); call != nil {
		switch call.Function {
		case "size":
			if len(call.Args) != 1 {
				return nil, errors.New("size() expects one argument")
			}
			arg, err := buildValueExpr(call.Args[0], schema)
			if err != nil {
				return nil, err
			}
			return &FunctionValue{
				Name: "size",
				Args: []ValueExpr{arg},
			}, nil
		case "now":
			return &LiteralValue{Value: timeNowUnix()}, nil
		case "_+_", "_-_", "_*_":
			value, ok, err := evaluateNumeric(expr)
			if err != nil {
				return nil, err
			}
			if ok {
				return &LiteralValue{Value: value}, nil
			}
		default:
			// Fall through to error return below
		}
	}

	return nil, errors.New("unsupported value expression")
}

func toComparisonOperator(fn string) (ComparisonOperator, error) {
	switch fn {
	case "_==_":
		return CompareEq, nil
	case "_!=_":
		return CompareNeq, nil
	case "_<_":
		return CompareLt, nil
	case "_>_":
		return CompareGt, nil
	case "_<=_":
		return CompareLte, nil
	case "_>=_":
		return CompareGte, nil
	default:
		return "", errors.Errorf("unsupported comparison operator %q", fn)
	}
}

func getIdentName(expr *exprv1.Expr) (string, error) {
	if ident := expr.GetIdentExpr(); ident != nil {
		return ident.GetName(), nil
	}
	return "", errors.New("expression is not an identifier")
}

func getConstValue(expr *exprv1.Expr) (interface{}, error) {
	v, ok := expr.ExprKind.(*exprv1.Expr_ConstExpr)
	if !ok {
		return nil, errors.New("expression is not a literal")
	}
	switch x := v.ConstExpr.ConstantKind.(type) {
	case *exprv1.Constant_StringValue:
		return v.ConstExpr.GetStringValue(), nil
	case *exprv1.Constant_Int64Value:
		return v.ConstExpr.GetInt64Value(), nil
	case *exprv1.Constant_Uint64Value:
		return int64(v.ConstExpr.GetUint64Value()), nil
	case *exprv1.Constant_DoubleValue:
		return v.ConstExpr.GetDoubleValue(), nil
	case *exprv1.Constant_BoolValue:
		return v.ConstExpr.GetBoolValue(), nil
	case *exprv1.Constant_NullValue:
		return nil, nil
	default:
		return nil, errors.Errorf("unsupported constant %T", x)
	}
}

func evaluateBool(call *exprv1.Expr_Call) (bool, bool, error) {
	val, ok, err := evaluateBoolExpr(&exprv1.Expr{ExprKind: &exprv1.Expr_CallExpr{CallExpr: call}})
	return val, ok, err
}

func evaluateBoolExpr(expr *exprv1.Expr) (bool, bool, error) {
	if literal, err := getConstValue(expr); err == nil {
		if b, ok := literal.(bool); ok {
			return b, true, nil
		}
		return false, false, nil
	}
	if call := expr.GetCallExpr(); call != nil && call.Function == "!_" {
		if len(call.Args) != 1 {
			return false, false, errors.New("NOT expects exactly one argument")
		}
		val, ok, err := evaluateBoolExpr(call.Args[0])
		if err != nil || !ok {
			return false, false, err
		}
		return !val, true, nil
	}
	return false, false, nil
}

func evaluateNumeric(expr *exprv1.Expr) (int64, bool, error) {
	if literal, err := getConstValue(expr); err == nil {
		switch v := literal.(type) {
		case int64:
			return v, true, nil
		case float64:
			return int64(v), true, nil
		}
		return 0, false, nil
	}

	call := expr.GetCallExpr()
	if call == nil {
		return 0, false, nil
	}

	switch call.Function {
	case "now":
		return timeNowUnix(), true, nil
	case "_+_", "_-_", "_*_":
		if len(call.Args) != 2 {
			return 0, false, errors.New("arithmetic requires two arguments")
		}
		left, ok, err := evaluateNumeric(call.Args[0])
		if err != nil {
			return 0, false, err
		}
		if !ok {
			return 0, false, nil
		}
		right, ok, err := evaluateNumeric(call.Args[1])
		if err != nil {
			return 0, false, err
		}
		if !ok {
			return 0, false, nil
		}
		switch call.Function {
		case "_+_":
			return left + right, true, nil
		case "_-_":
			return left - right, true, nil
		case "_*_":
			return left * right, true, nil
		default:
			return 0, false, errors.Errorf("unsupported arithmetic operator %q", call.Function)
		}
	default:
		return 0, false, nil
	}
}

func timeNowUnix() int64 {
	return time.Now().Unix()
}
