package main

func UNUSED(x ...interface{}) {}

func main() {
	p := new(Gitox)
	p.repoPath = "/Users/ameen/Desktop/utils/go/"
	p.setup()
	p.init()
	p.add("main.go")
	p.commit("yess daday")
	p.logs()
}
