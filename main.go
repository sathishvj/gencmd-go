package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	//"golang.design/x/clipboard"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var Version string = "to be set in ld flags"

const MaxCandidates int = 8

type Request struct {
	Prompt          PromptType `json:"prompt"`
	Temperature     float64    `json:"temperature"`
	TopK            int        `json:"top_k"`
	TopP            float32    `json:"top_p"`
	CandidateCount  int        `json:"candidate_count"`
	MaxOutputTokens int        `json:"max_output_tokens"`
	//StopSequences   []any   `json:"stop_sequences"`
	SafetySetting []SafetySetting `json:"safety_settings"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold int    `json:"threshold"`
}

type PromptType struct {
	Text string `json:"text"`
}

type Response struct {
	Candidates []Candidate `json:"candidates"`
}
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}
type Candidate struct {
	Output        string         `json:"output"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

type Args struct {
	os          string
	num         int
	temp        float64
	verbose     bool
	lines       bool
	help        bool
	cmd         string
	year        int
	warning     bool
	version     bool
	interactive bool
}

var args = Args{
	os:          "unix",
	num:         4,
	temp:        0.8,
	verbose:     false,
	lines:       false,
	help:        false,
	cmd:         "",
	year:        0,
	version:     false,
	warning:     false,
	interactive: false,
}

func showExamples() {
	// print usage
	fmt.Println(`
Examples:
	gencmd -n 5 -t 0.9 convert the first 10 seconds of an mp4 video into a gif
	gencmd -c grep find files that contain the text html
	gencmd -o windows find files that has extension pdf
	gencmd -i -c git recursively remove a directory from git but not from local

Written by Sathish VJ`)
}

func parseFlags() {
	args.os = runtime.GOOS
	if args.os == "darwin" {
		args.os = "unix"
	}

	flag.StringVar(&args.os, "o", args.os, "Operating system. Example: unix, linux, windows. Defaults to your OS.")
	flag.IntVar(&args.num, "n", args.num, "Max number of results to show. [1-8]. Default is 4. Suggestions are deduped; so you might get less than the number specified.")
	flag.Float64Var(&args.temp, "t", args.temp, "Temperature [0.0-1.0]. Default is 0.9.")
	flag.BoolVar(&args.verbose, "v", args.verbose, "Verbose. Default off.")
	flag.BoolVar(&args.lines, "l", args.lines, "Show line numbers. Default off.")
	flag.BoolVar(&args.help, "h", args.help, "Show usage.")
	flag.StringVar(&args.cmd, "c", args.cmd, "Command/Programme to use. Example: grep, ffmpeg, gcloud, curl. Default is empty.")
	flag.IntVar(&args.year, "y", args.year, "Year (included) post which the cmd is likely to have been used. This attempts to avoid older versions and options. Example: 2021, 2020, 2019. Default is none.")
	flag.BoolVar(&args.version, "version", false, "Show version of this build.")
	flag.BoolVar(&args.warning, "warning", false, "Suppress warning. Default off.")
	flag.BoolVar(&args.interactive, "i", false, "Copy command interactively. Default off.")

	flag.Parse()

	if args.num < 1 {
		args.num = 1
		fmt.Println("Number of suggestions cannot be less than 1. Setting it to 1.")
	}
	if args.num > MaxCandidates {
		args.num = MaxCandidates
		fmt.Println("Number of suggestions cannot be more than 10. Setting it to 10.")
	}
	if args.temp < 0.0 {
		args.temp = 0.0
		fmt.Println("Temperature cannot be less than 0.0. Setting it to 0.0.")
	}
	if args.temp > 1.0 {
		args.temp = 1.0
		fmt.Println("Temperature cannot be more than 1.0. Setting it to 1.0.")
	}
	if args.interactive {
		args.lines = true
	}

}

func checkAndGetAPIKey() string {
	// check if API_KEY is set in the environment variables
	value, exists := os.LookupEnv("GENCMD_API_KEY")
	if !exists {
		fmt.Println(`Error! GENCMD_API_KEY is not set in the environment variables.
Please set the API key in the environment variables.`)
		os.Exit(1)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		fmt.Println(`Error! GENCMD_API_KEY is not set in the environment variables.
Please set the API key in the environment variables.`)
		os.Exit(1)
	}
	return value
}

var (
	prompt_context = `Given a task, you should generate accurate, popular, precise, error free, effective ${OS} command line commands and options. ${CMD} ${YEAR} The task is:
`
	year_prompt = `The commands should be generated for any year from ${YEAR} to now.`
	cmd_prompt  = `Generate commands only for the ${CMD} command.`
)

func makeRequestString(args Args, userPrompt string) Request {
	prompt_context = strings.ReplaceAll(prompt_context, "${OS}", args.os)
	if args.cmd != "" {
		cmd_prompt = strings.ReplaceAll(cmd_prompt, "${CMD}", args.cmd)
		prompt_context = strings.ReplaceAll(prompt_context, "${CMD}", cmd_prompt)
	} else {
		prompt_context = strings.ReplaceAll(prompt_context, "${CMD}", "")
	}
	if args.year > 0 {
		year_prompt = strings.ReplaceAll(year_prompt, "${YEAR}", strconv.Itoa(args.year))
		prompt_context = strings.ReplaceAll(prompt_context, "${YEAR}", year_prompt)
	} else {
		prompt_context = strings.ReplaceAll(prompt_context, "${YEAR}", "")
	}
	prompt_context += userPrompt
	//fmt.Println("Full prompt context: ", prompt_context)

	r := Request{
		Temperature:     args.temp,
		TopK:            40,
		TopP:            0.95,
		CandidateCount:  MaxCandidates,
		MaxOutputTokens: 1024,
		//StopSequences: [],
		Prompt: PromptType{Text: prompt_context},
		SafetySetting: []SafetySetting{
			{
				Category:  "HARM_CATEGORY_DEROGATORY",
				Threshold: 1,
			},
			{
				Category:  "HARM_CATEGORY_TOXICITY",
				Threshold: 1,
			},
			{
				Category:  "HARM_CATEGORY_VIOLENCE",
				Threshold: 2,
			},
			{
				Category:  "HARM_CATEGORY_SEXUAL",
				Threshold: 2,
			},
			{
				Category:  "HARM_CATEGORY_MEDICAL",
				Threshold: 2,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS",
				Threshold: 2,
			},
		},
	}
	return r
}

func makeHTTPRequest(url string, reqData Request) Response {
	data, err := json.Marshal(reqData)
	if err != nil {
		fmt.Printf("Error occurred while converting request data to JSON: %s\n", err.Error())
		os.Exit(1)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error occurred while creating HTTP request: %s\n", err.Error())
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error occurred while making HTTP request: %s\n", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	//log.Printf("Response status code: %d\n", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error occurred while making HTTP request: %s\n", err.Error())
		os.Exit(1)
	}

	var respData Response
	err = json.Unmarshal(body, &respData)
	if err != nil {
		fmt.Printf("Error occurred while unmarshalling response: %s\n", err.Error())
		os.Exit(1)
	}

	return respData
}

func verbose(s string) {
	if args.verbose {
		fmt.Println(s)
	}
}

// I occasionally see random characters at the beginning and end of the commands. This function cleans them up.
func cleanCmd(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```\n") {
		s = strings.TrimPrefix(s, "```\n")
	}
	if strings.HasSuffix(s, "\n```") {
		s = strings.TrimSuffix(s, "\n```")
	}

	artifacts := []string{"```", "**"}
	for _, artifact := range artifacts {
		s = strings.TrimLeft(s, artifact)
		s = strings.TrimRight(s, artifact)
	}

	return s
}
func dedup(input []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func main() {
	parseFlags()
	if args.help {
		//usage()
		flag.Usage()
		showExamples()
		os.Exit(0)
	}

	if args.version {
		fmt.Println("gencmd Version: ", Version)
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		//usage()
		flag.Usage()
		showExamples()
		os.Exit(0)
	}

	// check if API_KEY is set in the environment variables. Exit if not set.
	key := checkAndGetAPIKey()

	reqData := makeRequestString(args, strings.Join(flag.Args(), " "))
	url := "https://generativelanguage.googleapis.com/v1beta2/models/text-bison-001:generateText?key=" + key

	resp := makeHTTPRequest(url, reqData)
	verbose(fmt.Sprintf("%v", resp))

	if !args.warning {
		fmt.Println("Warning! These suggestions are generated. They might not be accurate. If you are performing any file/folder/data destructive tasks, please back up your original data before trying it out.\n")
	}
	//fmt.Println("Suggestions:")

	suggestions := []string{}
	for _, candidate := range resp.Candidates {
		s := cleanCmd(candidate.Output)
		suggestions = append(suggestions, s)
	}
	suggestions = dedup(suggestions)

	if len(suggestions) < args.num {
		args.num = len(suggestions)
	}
	suggestions = suggestions[:args.num]

	for i, s := range suggestions {
		if args.lines {
			fmt.Println(fmt.Sprintf("%2d: %s", i+1, s))
		} else {
			fmt.Println(fmt.Sprintf("%s", s))
		}
	}

	if !args.interactive {
		os.Exit(0)
	}

	interactiveRun(suggestions)
}

/*
// ref: https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
// Splitting a string at Space, except inside quotation marks
func split(s string) []string {
	quoted := false
	a := strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' || r == '\'' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
	return a
}
*/

func interactiveRun(suggestions []string) {
choices:
	fmt.Print("\nNumber to copy to clipboard. q to quit. Enter your choice: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	input = strings.TrimSpace(input)

	if input == "q" {
		os.Exit(0)
	}

	opt, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Invalid choice. Neither q nor a number: %d\n", input)
		goto choices
	}
	if opt < 1 || opt > len(suggestions) {
		fmt.Printf("Invalid number choice. Must be within 1 and %d\n", len(suggestions))
		goto choices
	}

	err = copyToClipboard(suggestions[opt-1])
	if err != nil {
		fmt.Printf("Error copying to clipboard: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Copied suggestion %d to clipboard.\n", opt)
	os.Exit(0)
}
func copyToClipboard(s string) error {
	return clipboard.WriteAll(s)
}
