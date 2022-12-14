package models

type Deep2Model struct {
	Deep2Prop1 string `json:"deep2_prop1"`
	Deep2Prop2 string `json:"deep2_prop2,omitempty"`
}

type DeepModel struct {
	Deep2Model
	DeepProp1 string       `json:"deep_prop1"`
	DeepSlice []SliceModel `json:"deep_slice"`
}

type SliceModel struct {
	SliceProp1 string   `json:"slice_prop1"`
	SliceProp2 []*int32 `json:"slice_prop2,omitempty"`
	SliceProp3 *float64 `json:"slice_prop3,omitempty"`
}
