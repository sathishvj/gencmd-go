# gencmd
Generate cmd line arguments from the terminal itself. 

You're on your terminal and you're trying to recall the right command line arguments. What's your workflow then? Shift to your browser, run a search, click through multiple links, read through multiple pages, and then finally find the right command line arguments? That's distracting. With `gencmd` your workflow continues in the command line itself.

Uses Google's PaLM API. 

### Examples
These are generated. So you might see different results.
```	
gencmd -n 5 -t 0.9 -v convert the first 10 seconds of an mp4 video into a gif

Suggestions: 
ffmpeg -ss 00:00:10 -i input.mp4 -vcodec gif -pix_fmt rgb24 -t 00:00:10 output.gif
ffmpeg -ss 00:00:00 -i input.mp4 -t 00:00:10 -vf "scale=320:-1,fps=25,format=gif" output.gif
ffmpeg -ss 00:00:10 -i input.mp4 -vf "fps=10,scale=320:-1" output.gif
ffmpeg -ss 00:00:10 -i input.mp4 -vf scale=320:-1 -pix_fmt yuv420p output.gif
ffmpeg -i input.mp4 -vf "trim=0:10" output.gif
```

```
gencmd -c grep find txt files that contain the text hello
	
Suggestions:
grep -r hello *.txt
grep "hello" *.txt
grep -r hello *.txt
grep -i hello *.txt
```

```
gencmd -o windows -n 2 find files that has extension pdf

Suggestions:
dir /s *.pdf
dir /s *.pdf
	
```

```
gencmd -c git -n 2 recursively remove a directory from git but not from local

Suggestions:
git rm -r --cached <directory>
git rm --cached -r <directory>
```

### Requires PaLM API key
This works with Google's PaLM API. You need to have an API key for it to work. 
If you have access to MakerSuite, you can get it from there: [MakerSuite - Get API key](https://makersuite.google.com/app/apikey). 
As of now, access is limited by Google. So you might have to put yourself on the waitlist.

### Installation
If you know Go, you can download the source code or install it from `github.com/sathishvj/gencmd-go`.

Alternatively, here are the steps for unix based systems (Mac, Linux). The steps for Windows should be similar, but I haven't tested it.

#### 1. Download the binary for you OS and architecture from the releases directory: 

For example, if you are on Mac with Apple Silicon, then you can do:
```
wget https://github.com/sathishvj/gencmd-go/raw/main/releases/darwin-arm64/gencmd 
chmod +x gencmd 
./gencmd -h
```

Available builds right now are:
 - darwin-amd64/gencmd
 - darwin-arm4/gencmd
 - linux-amd64/gencmd
 - linux-arm644/gencmd
 - windows-amd64/gencmd.exe
 - windows-arm644/gencmd.exe


#### 2. Export the API key.
```
export GENCMD_API_KEY=<your api key>
```
You need to get this API key from [MakerSuite - Get API key](https://makersuite.google.com/app/apikey), if you have access.

#### 3. Run a basic command to test it out.
```
./gencmd -o unix -c grep find txt files that contain the text hello
```

You need to get this API key from [MakerSuite - Get API key](https://makersuite.google.com/app/apikey), if you have access.

#### 4. More Permanent Options
 - You can add the binary to your path.
 - You can add the export command for the API key to your .bashrc or .zshrc or .profile file.

### Options
```
 -c Command/Programme to use. Example: grep, ffmpeg, gcloud, curl. Default is empty.
 -l	Show line numbers. Default off.
 -n Number of results to generate. Max 10. Default is 4. (default 4)
 -o Operating system. Example: unix, linux, windows (default "unix")
 -t Temperature [0.0-1.0]. Default is 0.9. (default 0.9)
 -y Year (included) post which the cmd is likely to have been used. This attempts to avoid older versions and options. Example: 2021, 2020, 2019. Default is none.
 
 -h	Show usage.
 -v	Verbose. Default off.
 -version Show version of this build.
```

### Warning!
Google's policies say they can use your data. So don't use this with sensitive data.

## License
MIT License. 

