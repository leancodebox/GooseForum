package setting

const storage = "./storage/"
const tmp = "./storage/tmp/"

func GetStorage() string {
	return storage
}

func GetTmp() string {
	return tmp
}
