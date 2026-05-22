# Cover Letter Writer

Turn clipboard text into a formatted PDF cover letter. No typing into templates, no copy-paste into Word.

![Example cover letter](screenshot.png)

## What it does

1. Reads your contact info from a file
2. Reads cover letter text from your clipboard
3. Saves a styled PDF to your Downloads folder

## Setup

Copy `user.json.example` to `user.json` and fill in your details:

```
cp user.json.example user.json
```

Required fields: `name`, `address`, `email`, `phone`.

Optional fields:
- `outputDir` — where to save the PDF. Defaults to your Downloads folder.
- `filename` — what to name the file. Defaults to `<Name>NoSpacesCoverLetter.pdf`.

## How to use

1. Write or paste your cover letter body into any app (Notes, ChatGPT, etc)
2. Copy it to your clipboard
3. Run the app:
   - Double-click `cover-letter-writter` if you have the downloaded version
   - Or open Terminal, go to this folder, and run: `go run .`
4. Find the PDF in your Downloads folder (or wherever you set `outputDir`)

Done.
