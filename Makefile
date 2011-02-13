ftpclient:ftp.8 main.8
	8l -o ftpclient main.8

ftp.8:ftp.go
	8g -o ftp.8 ftp.go

main.8:main.go
	8g -o main.8 main.go

clean:
	rm -rf ftp.8 main.8 ftpclient	
