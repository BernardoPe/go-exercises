package main

import (
	"context"
	"file-downloader/internal"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	urlsFlag := flag.String("urls", "", "Comma-separated list of URLs to download")
	dirFlag := flag.String("dir", "./downloads", "Directory to save downloaded files")
	flag.Parse()

	if *urlsFlag == "" {
		log.Fatal("Error: -urls flag is required\nUsage: go run main.go -urls=url1,url2,... [-dir=download_directory]")
	}

	urls := strings.Split(*urlsFlag, ",")
	directory := *dirFlag

	fmt.Printf("Preparing to download %d file(s) to %s\n\n", len(urls), directory)

	downloads, err := internal.PrepareDownloads(urls, directory)
	if err != nil {
		log.Fatalf("Error preparing downloads: %v", err)
	}

	for i, d := range downloads {
		fmt.Printf("[%d] %s\n", i+1, d.FilePath)
	}
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, cancelling downloads...")
		cancel()
	}()

	var wg sync.WaitGroup

	internal.StartProgressListener(downloads, &wg)

	err = internal.StartAll(ctx, downloads, &wg)
	if err != nil {
		log.Fatalf("Error starting downloads: %v", err)
	}

	wg.Wait()

	var downloadErrors []error
	for i, d := range downloads {
		if err := d.Err(); err != nil {
			downloadErrors = append(downloadErrors, fmt.Errorf("[%d] %s: %w", i+1, d.URL, err))
		}
	}

	if len(downloadErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d download(s) failed:\n", len(downloadErrors))
		for _, err := range downloadErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("\nAll downloads completed successfully!")
}
