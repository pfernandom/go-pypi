.PHONY: downloadFile downloadDescriptor intTest

downloadFile:
	wget -O numpy-2.3.4.tar.gz "http://localhost:4040/proxy/packages/b5/f4/098d2270d52b41f1bd7db9fc288aaa0400cb48c2a3e2af6fa365d9720947/numpy-2.3.4.tar.gz?originalHost=pypi.org&originalScheme=https" && \
	file numpy-2.3.4.tar.gz && \
	rm numpy-2.3.4.tar.gz

downloadDescriptor:
	curl -v "http://localhost:4040/simple/numpy/" | jq .

intTest: downloadFile downloadDescriptor
	