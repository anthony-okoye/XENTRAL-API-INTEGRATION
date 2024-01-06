package request

type Request struct {
	Entity   string         `json:"entity"`
	Data     map[string]any `json:"data"`
	Metadata Metadata       `json:"metadata"`
}

type GetRequest struct {
	Entity   string   `json:"entity"`
	Data     Data     `json:"data"`
	Metadata Metadata `json:"metadata"`
}

type Data struct {
	ID             string `json:"id"`
	SalesChannelID string `json:"sales_channel_id"`
}

type Metadata struct {
	Filter           Filter         `json:"filter"`
	OrderBy          OrderBy        `json:"orderBy"`
	Fields           []string       `json:"fields"`
	Relationships    []Relationship `json:"relationships"`
	Limit            int            `json:"limit"`
	Offset           int            `json:"offset"`
	OverrideOnUpdate []string       `json:"override_on_update"`
	UpdateFields     []string       `json:"update_fields"`
}

type Relationship struct {
	Name           string        `json:"name"`
	RelationParams []FilterParam `json:"relation_params"`
}

type Response struct {
	Status     bool     `json:"status"`
	Errors     []string `json:"errors"`
	Data       any      `json:"data,omitempty"`
	NextOffset int      `json:"next_offset,omitempty"`
	Total      int      `json:"total,omitempty"`
}

type Filter struct {
	Should []FilterParam `json:"should,omitempty"`
	Must   []FilterParam `json:"must,omitempty"`
}

type FilterParam struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
	Type  string `json:"type"`
}

type ConditionStatement struct {
	Main   string        `json:"main"`
	Values []interface{} `json:"condition.Values"`
}

type OrderBy struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type SumBy struct {
	Key string `json:"key"`
	As  string `json:"as"`
}
