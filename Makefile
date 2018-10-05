test:
	@rm -rf test_logrotate
	@# For check file exists test
	@mkdir -p test_logrotate
	@touch test_logrotate/test.log
	@touch test_logrotate/test-1.log
	@# For file size test
	@echo "hello world!" > test_logrotate/get-file-size-test.log
	go test -v
