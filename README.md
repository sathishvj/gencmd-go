# gencmd
Generate cmd line arguments from the cmd line itself. 
Uses Google's PaLM API. 

### Warning! 
Google's policies say they can use your data. So don't use this with sensitive data.

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

## Required! PaLM API key
This works with Google's PaLM API. You need to have an API key for it to work. 
If you have access to MakerSuite, you can get it from there: [MakerSuite - Get API key](https://makersuite.google.com/app/apikey)

## Installation
If you know Go, you can install it from `github.com/sathishvj/gencmd-go`.

Alternatively, here are the steps:
1. Download the binary for you OS and architecture from the releases directory: 

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
