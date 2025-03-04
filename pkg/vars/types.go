package vars

// VarType 定义变量类型
type VarType int

// 变量类型常量
const (
	VarTypeString VarType = iota
	VarTypeInt
	VarTypeFloat
	VarTypeBool
	VarTypeList
	VarTypeMap
)

// Variable 定义变量接口
type Variable interface {
	GetName() string
	GetValue() interface{}
	GetType() VarType
	SetValue(value interface{}) error
}

// BaseVariable 基础变量实现
type BaseVariable struct {
	Name  string
	Value interface{}
	Type  VarType
}

// GetName 获取变量名
func (v *BaseVariable) GetName() string {
	return v.Name
}

// GetValue 获取变量值
func (v *BaseVariable) GetValue() interface{} {
	return v.Value
}

// GetType 获取变量类型
func (v *BaseVariable) GetType() VarType {
	return v.Type
}

// SetValue 设置变量值
func (v *BaseVariable) SetValue(value interface{}) error {
	v.Value = value
	return nil
}

// NewVariable 创建新变量
func NewVariable(name string, value interface{}, varType VarType) Variable {
	return &BaseVariable{
		Name:  name,
		Value: value,
		Type:  varType,
	}
}