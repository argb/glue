package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store map[string]Symbol
	numDefinitions int

	FreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	var free []Symbol

	return &SymbolTable{store: s, FreeSymbols: free}
}

// NewEnclosedSymbolTable /**
/*
构造作用于链
 */
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer

	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: GlobalScope}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	}else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++

	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {

	obj, ok := s.store[name]
	if !ok { // 取不到
		if s.Outer != nil{ // 并且还可以继续往上找
			//如果在外层取到了，判断其是否是全局Symbol还是Builtin
			obj, ok = s.Outer.Resolve(name)
			if ok {
				if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
					// 如果是全局的或者builtin的，直接返回
					return obj, ok
				}else{
					// 否则，将其设置为自由变量(free variable)。 一个变量，它存在，但不是本地的，又不是上面两种特殊情况的，就把它看做自由的
					// 然后将其绑定到自己的当前符号表中（相当于把这个符号复制了一份，修改下属性，再存起来），这样当前作用域就一直持有这个自由变量了，
					free := s.defineFree(obj)
					return free, ok
				}
			}else {
				return s.Outer.Resolve(name)
			}

		}else { // 并且无法继续找了
			return Symbol{}, false
		}
	}else {
		// 默认取到了直接返回了
		return obj, ok
	}

}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Name: name,
		Index: index,
		Scope: BuiltinScope,
	}
	s.store[name] = symbol

	return symbol
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols)-1}
	symbol.Scope = FreeScope

	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope} // Index 值随便写
	s.store[name] = symbol
	return symbol
}
