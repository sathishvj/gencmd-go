package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var Version string = "to be set in ld flags"

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
	os      string
	num     int
	temp    float64
	verbose bool
	lines   bool
	help    bool
	cmd     string
	year    int
	version bool
}

var args = Args{
	os:      "unix",
	num:     4,
	temp:    0.9,
	verbose: false,
	lines:   false,
	help:    false,
	cmd:     "",
	year:    0,
	version: false,
}

func showExamples() {
	// print usage
	fmt.Println(`
Examples:
	gencmd -n 5 -t 0.9 -v convert the first 10 seconds of an mp4 video into a gif
	gencmd -c grep find files that contain the text html
	gencmd -o windows find files that has extension pdf
	gencmd -c git recursively remove a directory from git but not from local

Written by Sathish VJ`)
}

func parseFlags() {
	// capture command line arguments os as string for os, n as int for number, t as float32 for temperature, v as bool for verbose, l as bool for lines

	/*
		osArg := flag.String("o", args.os, "Operating system. Example: unix, linux, windows")
		numArg := flag.Int("n", args.num, "Number of results to generate. Max 10. Default is 4.")
		tempArg := flag.Float64("t", float64(args.temp), "Temperature [0.0-1.0]. Default is 0.9.")
		verboseArg := flag.Bool("v", args.verbose, "Verbose. Default off.")
		linesArg := flag.Bool("l", args.lines, "Show line numbers. Default off.")
		helpArg := flag.Bool("h", args.help, "Show usage.")
		cmdArg := flag.String("c", args.cmd, "Command/Programme to use. Example: grep, ffmpeg, gcloud, curl. Default is empty.")
		yearArg := flag.Int("y", args.year, "Year (included) post which the cmd is likely to have been used. This attempts to avoid older versions and options. Example: 2021, 2020, 2019. Default is none.")
		versionArg := flag.Bool("version", false, "Show version of this build.")
	*/

	flag.StringVar(&args.os, "o", args.os, "Operating system. Example: unix, linux, windows")
	flag.IntVar(&args.num, "n", args.num, "Number of results to generate. Max 10. Default is 4.")
	flag.Float64Var(&args.temp, "t", args.temp, "Temperature [0.0-1.0]. Default is 0.9.")
	flag.BoolVar(&args.verbose, "v", args.verbose, "Verbose. Default off.")
	flag.BoolVar(&args.lines, "l", args.lines, "Show line numbers. Default off.")
	flag.BoolVar(&args.help, "h", args.help, "Show usage.")
	flag.StringVar(&args.cmd, "c", args.cmd, "Command/Programme to use. Example: grep, ffmpeg, gcloud, curl. Default is empty.")
	flag.IntVar(&args.year, "y", args.year, "Year (included) post which the cmd is likely to have been used. This attempts to avoid older versions and options. Example: 2021, 2020, 2019. Default is none.")
	flag.BoolVar(&args.version, "version", false, "Show version of this build.")

	flag.Parse()

	/*
		// these cannot be null since we are giving defaults, right? Check later
		if osArg != nil {
			args.os = *osArg
		}
		if numArg != nil {
			args.num = *numArg
		}
		if tempArg != nil {
			args.temp = float32(*tempArg)
		}
		if verboseArg != nil {
			args.verbose = *verboseArg
		}
		if linesArg != nil {
			args.lines = *linesArg
		}
		if helpArg != nil {
			args.help = *helpArg
		}
		if cmdArg != nil {
			args.cmd = *cmdArg
		}
		if yearArg != nil {
			args.year = *yearArg
		}
		if versionArg != nil {
			args.version = *versionArg
		}
	*/

	if args.num < 1 {
		args.num = 1
		fmt.Println("Number of suggestions cannot be less than 1. Setting it to 1.")
	}
	if args.num > 10 {
		args.num = 10
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
	prompt_context = `Given a task, you should generate very good at generating accurate, popular, precise, error free, effective ${OS} command line commands and options. ${CMD} ${YEAR} The task is:
`
	year_prompt = `The commands should be generated for any year from ${YEAR} to now.`
	cmd_prompt  = `Generate commands only for the ${CMD} command.`
)

func makeRequestString(key string, args Args, user_prompt string) Request {
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
	prompt_context += user_prompt
	//fmt.Println("Full prompt context: ", prompt_context)

	r := Request{
		Temperature:     args.temp,
		TopK:            40,
		TopP:            0.95,
		CandidateCount:  args.num,
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
		log.Fatalf("Error occurred while converting request data to JSON: %s", err.Error())
		os.Exit(1)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error occurred while creating HTTP request: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error occurred while making HTTP request: %s", err.Error())
	}
	defer resp.Body.Close()

	//log.Printf("Response status code: %d\n", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error occurred while making HTTP request: %s", err.Error())
	}

	var respData Response
	err = json.Unmarshal(body, &respData)
	if err != nil {
		log.Fatalf("Error occurred while unmarshalling response: %s", err.Error())
	}

	return respData
}

func verbose(s string) {
	if args.verbose {
		fmt.Println(s)
	}
}

func cleanCmd(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```\n") {
		s = strings.TrimPrefix(s, "```\n")
	}
	if strings.HasSuffix(s, "\n```") {
		s = strings.TrimSuffix(s, "\n```")
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return s
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
	reqData := makeRequestString(key, args, strings.Join(flag.Args(), " "))
	url := "https://generativelanguage.googleapis.com/v1beta2/models/text-bison-001:generateText?key=" + key

	resp := makeHTTPRequest(url, reqData)
	verbose(fmt.Sprintf("%v", resp))

	fmt.Println("\nThese suggestions are generated. They might not be accurate. If you are performing any file/folder/data destructive tasks, please back up your original data before trying it out.\n")
	fmt.Println("Suggestions:")
	for i, candidate := range resp.Candidates {
		s := cleanCmd(candidate.Output)
		if args.lines {
			fmt.Printf("%d: %s\n", i+1, s)
		} else {
			fmt.Printf("%s\n", s)
		}
	}
}
