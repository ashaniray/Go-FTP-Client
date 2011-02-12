ftpclient:ftp.8
	8l -o ftpclient ftp.8

ftp.8:ftp.go
	8g -o ftp.8 ftp.go
	
clean:
	rm -rf ftp.8 ftpclient	
