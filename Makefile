COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out
COVERAGEREPORT=$(COVERDIR)/report.html

test:
	@ginkgo --failFast ./...

test-watch:
	@ginkgo watch -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@ginkgo -r -covermode=count --cover --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find . -type f -name '*.coverprofile' -exec cat {} \; -exec rm -f {} \; | grep -h -v "^mode:" >> ${COVERAGEFILE}

coverage: coverage-ci
	@sed -i -e "s|_$(PROJECT_ROOT)/|./|g" "${COVERAGEFILE}"
	@cp "${COVERAGEFILE}" coverage.txt

coverage-html:
	@go tool cover -html="${COVERAGEFILE}" -o $(COVERAGEREPORT)
	@xdg-open $(COVERAGEREPORT) 2> /dev/null > /dev/null

vet:
	@go vet ./...

fmt:
	@go fmt ./...

.PHONY: test test-watch coverage coverage-ci coverage-html vet fmt
