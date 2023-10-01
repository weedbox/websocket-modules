package websocket_server

type Metadata struct {
	entries map[string]interface{}
}

func NewMetadata() *Metadata {
	return &Metadata{
		entries: make(map[string]interface{}),
	}
}

func (md *Metadata) Delete(key string) {
	delete(md.entries, key)
}

func (md *Metadata) Get(key string) interface{} {
	return md.entries[key]
}

func (md *Metadata) Set(key string, value interface{}) {
	md.entries[key] = value
}

func (md *Metadata) GetString(key string) string {
	val, ok := md.entries[key]
	if !ok {
		return ""
	}
	return val.(string)
}

func (md *Metadata) GetInt(key string) int {
	val, ok := md.entries[key]
	if !ok {
		return 0
	}
	return val.(int)
}
