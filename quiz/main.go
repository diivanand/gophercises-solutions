package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"quiz/lib"
	"time"
)

func main() {
	inputFile := flag.String("in", "problems.csv", "quiz input CSV file")
	quizTimeout := flag.Duration("timeout", 30*time.Second, "quiz timeout")
	shouldShuffleQuestions := flag.Bool("shuffleQuestions", false, "shuffle the order of quiz questions")
	flag.Parse()
	quizCmdLineArgs := lib.QuizCmdLineArgs{
		TestInputFile:          *inputFile,
		QuizTimeout:            *quizTimeout,
		ShouldShuffleQuestions: *shouldShuffleQuestions,
	}
	results, err := lib.Run(quizCmdLineArgs, os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("quiz failed with error: %v", err)
	}
	fmt.Printf(
		"Quiz results: %d/%d correct (%d incorrect)",
		results.NumCorrect,
		results.TotalNumberOfQuestions,
		results.NumIncorrect,
	)
	fmt.Println()
	fmt.Println("Quiz completed")
}
