package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ttacon/chalk"
)

type Gitox struct {
	repoPath    string
	opjectsPath string
	headPath    string
	indexPath   string
}

func (p *Gitox) setup() {
	p.repoPath += ".gitox/"
	p.opjectsPath = p.repoPath + "objects/"
	p.headPath = p.repoPath + "head"
	p.indexPath = p.repoPath + "index"
}

func (git Gitox) init() {
	os.MkdirAll(git.repoPath, 0755)
	os.MkdirAll(git.opjectsPath, 0755)
	if _, err := os.Stat(git.headPath); os.IsNotExist(err) {
		headFile, err := os.Create(git.headPath)
		if err != nil {
			fmt.Println("Error creating index file:", err)
			return
		}
		headFile.Close()
	}
	if _, err := os.Stat(git.indexPath); os.IsNotExist(err) {
		indexFile, err := os.Create(git.indexPath)
		if err != nil {
			fmt.Println("Error creating index file:", err)
			return
		}
		indexFile.Close()
	}
}

func (p Gitox) hash(content []byte) string {
	h := sha1.New()
	h.Write(content)
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}

func (p Gitox) add(fileToBeAdd string) {
	data, err := os.ReadFile(fileToBeAdd)
	if err != nil {
		panic(err)
	}

	hash := p.hash(data)
	newFileHash := p.opjectsPath + hash
	f, err := os.OpenFile(newFileHash, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("fild to open or create the file ", newFileHash)
		return
	}
	f.WriteString(string(data))
	p.updateStaging(fileToBeAdd, hash)
	f.Close()
}

type Stage struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func (p Gitox) updateStaging(filepath string, filehash string) {
	UNUSED(filepath, filehash) //tmp
	data, err := os.ReadFile(p.indexPath)
	if err != nil {
		print("\nerr to open index file -> ", err, "\n")
	}
	var stages []Stage
	json.Unmarshal(data, &stages)
	stg := Stage{
		Path: filepath,
		Hash: filehash,
	}
	stages = append(stages, stg)
	newData, err := json.Marshal(stages)
	if err != nil {
		fmt.Print("faild to convers stage to byt")
		return
	}
	f, err2 := os.OpenFile(p.indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err2 != nil {
		fmt.Print("faild to open file ")
		return
	}
	f.WriteString(string(newData))
	f.Close()
}

type CommitData struct {
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
	Files   []Stage   `json:"files"`
	Parrent string    `json:"parrent"`
}

func (git Gitox) commit(message string) {
	UNUSED(message) //tmp
	index, err := os.ReadFile(git.indexPath)
	if err != nil {
		print("\nerr to open index file -> ", err, "\n")
	}
	var stages []Stage
	json.Unmarshal(index, &stages)
	parrentHead := git.getCurrentHead()

	commit := CommitData{
		Message: message,
		Date:    time.Now(),
		Files:   stages,
		Parrent: parrentHead,
	}
	data, err2 := json.Marshal(commit)
	if err2 != nil {
		fmt.Print("faild to convert commit to byte")
	}
	commitHash := git.hash(data)
	commitPath := git.opjectsPath + commitHash
	f, err := os.OpenFile(commitPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("fild to open or create the file ", commitPath)
		return
	}

	f.WriteString(string(data))
	f.Close()
	headF, headErr := os.OpenFile(git.headPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if headErr != nil {
		fmt.Print("faild to open the file ", git.headPath)
		return
	}
	headF.Truncate(0)
	headF.WriteString(commitHash)
	headF.Close()
	//cearing the staging
	indexF, indexErr := os.OpenFile(git.indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if indexErr != nil {
		fmt.Print("faild to open the file ", git.indexPath)
		return
	}

	indexF.Truncate(0)
	indexF.Close()
}

func (git Gitox) getCurrentHead() string {
	data, err := os.ReadFile(git.headPath)
	if err != nil {
		print("\nerr to open index file -> ", err, "\n")
	}
	ret := string(data)
	return ret
}

// stoped here it need to show logs when i get to show logs like git

func (git Gitox) logs() {
	currentCommitHash := git.getCurrentHead()
	for len(currentCommitHash) > 0 {
		commit := git.getCommitData(currentCommitHash)
		git.printCommit(commit, currentCommitHash)
		currentCommitHash = commit.Parrent
	}
}

// return commited data based on the gived hash, return as a json formated by type CommitData
func (git Gitox) getCommitData(commitHash string) CommitData {
	commitDataBytes, err := os.ReadFile(git.opjectsPath + commitHash)
	var commit CommitData
	if err != nil {
		fmt.Print("fiald to read the file : ", git.opjectsPath+commitHash)
		return commit
	}

	err = json.Unmarshal(commitDataBytes, &commit)
	if (err == &json.InvalidUnmarshalError{}) {
		fmt.Print("faild to pars commits to json please try again later\n")
		return commit
	}
	return commit
}

// return all commited files in one string sep by "-"
func (git Gitox) getCommitPaths(files []Stage) string {
	var fullstring string
	for i := 0; i < len(files); i++ {
		if i+1 == len(files) {
			fullstring += files[i].Path
		} else {
			fullstring += files[i].Path + "  -  "
		}
	}
	return fullstring
}

func (git Gitox) printCommit(commit CommitData, hash string) {
	commitPaths := git.getCommitPaths(commit.Files)
	fmt.Println(
		chalk.Yellow,
		chalk.Bold.TextStyle("Commit : "),
		chalk.Bold.TextStyle(hash+"\n"),
		chalk.Reset,
		"\n",
		chalk.Bold.TextStyle("Message: "),
		commit.Message+"\n",
		chalk.Bold.TextStyle("Date: "),
		commit.Date,
		"\n",
		chalk.Bold.TextStyle("Files: "),
		commitPaths,
	)
	fmt.Print("\n___________________________\n")
}

func (git Gitox) diff(leftFile string, rightFile string) {
	//will develop soon
}
