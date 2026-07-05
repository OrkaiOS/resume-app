.PHONY: run install dev build clean

build:
	cd frontend && npm run build
	mkdir -p backend/cmd/static
	cp -R frontend/dist/* backend/cmd/static/
	cd backend && go build -tags prod -o bin/resume-app ./cmd

run: build
	./backend/bin/resume-app

install: build
	@if [ -n "$$GOPATH" ]; then \
		cp backend/bin/resume-app "$$GOPATH/bin/orkai-resume"; \
	else \
		cp backend/bin/resume-app /usr/local/bin/orkai-resume; \
	fi

dev:
	@trap 'kill 0' INT TERM; \
	(cd frontend && npm run dev) & \
	(cd backend && go run ./cmd) & \
	wait

clean:
	rm -rf backend/bin
	find backend/cmd/static -mindepth 1 -not -name .gitkeep -delete 2>/dev/null || true
	rm -rf frontend/dist