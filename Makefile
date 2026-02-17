.PHONY: repomix question clean

questions:
	go build -o bin/questions internal/questions/main.go && chmod u+x bin/questions

rebuild: clean questions

repomix:
	rm -f voting-app.json && pnpm dlx repomix

clean :
	rm -rf bin/ voting.db