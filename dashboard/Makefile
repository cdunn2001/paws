#
# This cmake code creates a target (libPaWsDashboard) that
# contains staticly linked text files, such as the REST API markdown,
# and a tiny wrapper to access these text files as std::strings.
#

dashboard.o: dashboard.html
	ld -r -b binary -o dashboard1.o dashboard.html
	objcopy --rename-section .data=.rodata,alloc,load,readonly,data,contents dashboard1.o $@

clean:
	rm -f dashboard.o dashboard1.o

stuff:
