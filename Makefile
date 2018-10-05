test:
	@# For check file exists test
	@mkdir -p test_logrotate
	@rm -rf test_logrotate/*.log
	@touch test_logrotate/test-empty.log
	@touch test_logrotate/test-empty-1.log
	@touch test_logrotate/test-existing.log
	@touch test_logrotate/test-existing-rotate.log
	@# For file size test
	@echo "hello world!" > test_logrotate/get-file-size-test.log
	go test -v $(TEST_ARGS)
