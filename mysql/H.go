package mysql

type Any = interface{}

type H = map[string]Any

func NewH(m map[string]interface{}) H {
	h := H{}
	for k, v := range m {
		h[k] = v
	}
	return h
}
