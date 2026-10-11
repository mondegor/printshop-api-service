-include Makefile.mk

deps:
	mrcmd go-dev deps

deps-upgrade:
	mrcmd go-dev get -u ./...
	mrcmd go-dev tidy

generate:
	mrcmd go-dev generate

lint:
	mrcmd go-dev fmt
	mrcmd go-dev lint

test:
	mrcmd go-dev test

test-report:
	mrcmd go-dev test-report

.PHONY: deps deps-upgrade generate lint test test-report
