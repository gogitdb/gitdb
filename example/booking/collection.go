package booking

//implement sort interface to allow us sort an array of files i.e []os.FileInfo
// type collection []*Model

// func (c collection) Len() int {
// 	return len(c)
// }
// func (c collection) Less(i, j int) bool {
// 	return c[i].CreatedAt.Before(c[j].CreatedAt)
// }
// func (c collection) Swap(i, j int) {
// 	c[i], c[j] = c[j], c[i]
// }

// type Collection struct {
// 	SortFunc func(i int, j int) bool
// 	Data     []*db.Model
// }

// func NewCollection(data []*db.Model) *Collection {
// 	return &Collection{Data: data}
// }

// func (c *Collection) Len() int {
// 	return len(c.Data)
// }
// func (c *Collection) Less(i, j int) bool {
// 	return c.SortFunc(i, j)
// }
// func (c *Collection) Swap(i, j int) {
// 	c.Data[i], c.Data[j] = c.Data[j], c.Data[i]
// }

// func (c *Collection) SortByDate(i, j int) bool {
// 	return c.Data[i].CreatedAt.Before(c.Data[j].CreatedAt)
// }

// func (c *Collection) SortByName(i, j int) bool {
// 	return c.Data[i].Name.Before(c.Data[j].CreatedAt)
// }

// c := NewCollection(a)
// c.SortFunc = c.SortByDate

// sort(c)
