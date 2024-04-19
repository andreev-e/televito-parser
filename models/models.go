package Models

type Add struct {
	Id           int
	User_id      int
	Status       int
	Location_id  uint16
	Name         string
	Description  string
	Price        int
	Price_usd    float32
	Source_class string
	Source_id    uint32
	CategoryId   uint16
	Approved     int
	Images       string
	Currency     string
	Updated_at   string
	Created_at   string
	Deleted_at   *string
}

type Category struct {
	Id         uint16
	Name       string
	ParentId   uint16
	Created_at string
	Updated_at string
	Deleted_at string
	Adds_count uint32
}

type User struct {
	Id          int
	Contact     interface{}
	Lang        string
	Currency    string
	Location_id uint16
	Timezone    string
	Created_at  string
	Updated_at  string
}
