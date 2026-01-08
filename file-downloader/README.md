# File Downloader

This is a simple program that downloads files from the given
URLs and saves them to the specified directory.

This program was used to gain familiarity with the 
use of goroutines and channels in Go for concurrent programming.

## Usage
To run the file downloader, use the following command:

```bash
go run main.go -urls=url1,url2,... -dir=download_directory
```

## Design

The file downloader uses goroutines and channels for concurrent file downloads. Each file downloads in its own goroutine. File writing happens simultaneously with network reading in the same goroutine. Progress updates are sent through channels. A single WaitGroup passed from main() tracks all goroutines including downloads and UI.

### Components

`FileDownload` (`internal/downloader.go`)

The `Prepare()` method validates URLs, creates HTTP requests, and initializes channels. The `Start()` method begins the download, writes to disk, and sends progress updates via the `LoadedBytes` channel.

Progress UI (`internal/ui.go`)

One listener goroutine runs per download, consuming from the `LoadedBytes` channels to track progress. A display updater goroutine refreshes all progress bars together every 1 second using ANSI cursor positioning. Speed calculation tracks download rate and estimates time remaining.

Main (`main.go`)

The `main()` function first validates URLs and creates HTTP connections without transferring data. It then initializes UI listeners and creates a `WaitGroup`. All downloads start, with each goroutine downloading and writing simultaneously. The `main()` function calls `wg.Wait()` to block until all downloads and UI updates complete. Finally, it displays any errors or confirms success.

### Concurrency Pattern
```
main()
 └─ WaitGroup ─┬─ Download 1 (HTTP read → File write + Progress updates)
               ├─ Download 2 (HTTP read → File write + Progress updates)
               ├─ Download N (HTTP read → File write + Progress updates)
               ├─ Progress Listener 1 (reads LoadedBytes channel)
               ├─ Progress Listener 2 (reads LoadedBytes channel)
               ├─ Progress Listener N (reads LoadedBytes channel)
               └─ Display Updater (refreshes all progress bars)
```