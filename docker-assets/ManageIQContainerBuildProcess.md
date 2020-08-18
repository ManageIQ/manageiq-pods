## Building and running a ManageIQ Container 

The following steps document the process for building a ManageIQ Container process with IBM Power Virtual Server Cloud

1. Configure the bundle with the github account information
    ```
    bundle config GITHUB__IBM__COM Kuldip-Nanda:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX --global
    bundle confg
     Settings are listed in order of priority. The top value will be used.
     github.ibm.com
     Set for the current user (/root/.bundle/config): "Kuldip-Nanda:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    ```
   Add the following lines in the ManageIQ Gemfile.<span style="color:red">
	```
	group :ibm_cloud_virtual_servers, :manageiq_default do
  	  gem 'manageiq-providers-ibm_cloud',  :git => 'https://manageiq/manageiq-providers-ibm_cloud.git',  :branch => "master"
	end
	```
	
   > *NOTE:* As the manageiq-providers-ibm_cloud.git is a internal repo, it is not possible to download this like any public repo. The only way is to either do https/ssh credential setup or generate a personal access token for cloning. The [personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) with full repo access can be setup in github.ibm.com page as:

			settings > developer settings > personal access tokens > generate new token'

	This will allow the bundler to fetch and install the ***manageiq_provider-ibm_cloud*** gemfile when doing ***bundle install*** or ***bin/setup***. 
	Add, commit and push these to remote.

2.	Create the manageiq rpm
	- Clone the public ***[manageiq-rpm_build](https://github.com/ManageIQ/manageiq-rpm_build)** repo
		> git clone https://github.com/ManageIQ/manageiq-rpm_build.git 
	- Make changes to the options.xml file in the config directory
		```
		diff --git a/config/options.yml b/config/options.yml
			index b7d7b15..3f9be99 100644
			--- a/config/options.yml
			+++ b/config/options.yml
			@@ -3,8 +3,8 @@ product_name:      manageiq
 			repos:
			   ref:             master
			   manageiq:
			-    url:           https://github.com/ManageIQ/manageiq.git
			-    ref:
			+    url:           https://Kuldip-Nanda:4d16662b66968eea1e4c41dbfb21d0e372031469@github.ibm.com/Kuldip-Nanda/			manageiq.git
			+    ref:           ibm
		   manageiq_appliance:
		     url:           https://github.com/ManageIQ/manageiq-appliance.git
		     ref:
		```
		>*NOTE:* Intead of changing the config/options.yml file one can also pass custom options. 

	- Build the docker container.
		In centos 8 intead of docker, podman can be used to build  and run container images 
		
		> [root@localhost manageiq-rpm_build]# podman build --tag manageiq-rpm:1.0 .

	- Run the *manageiq-rpm:1.0* container image. This will boot to the development shell from where one can start the build process
		```
		podman run -it -v $PWD/output:/output manageiq-rpm:latest
        ~~~
        [root@d50306c16c4b build_scripts]# bin/build.rb
		 ---> ManageIQ::RPMBuild::SetupSourceRepos#initialize
		~~~
		 ---> git clone --depth 1 -b ibm https://Kuldip-Nanda:XXXXXXXXXXXXXXXXXXXXXXXXXXX@github.ibm.com/Kuldip-Nanda/manageiq.git manageiq
		Cloning into 'manageiq'...
		remote: Enumerating objects: 3243, done.
		remote: Counting objects: 100% (3243/3243), done.

		```

	- Once the RPMs are build we can copy these to mounted volume  *output* 

      ```
		[root@d50306c16c4b build_scripts]# ls /root/BUILD/rpms/x86_64
		manageiq-appliance-11.0.0-0.1.20200807153703.el8.x86_64.rpm        
		manageiq-core-11.0.0-0.1.20200807153703.el8.x86_64.rpm   
		manageiq-pods-11.0.0-0.1.20200807153703.el8.x86_64.rpm    
		manageiq-ui-11.0.0-0.1.20200807153703.el8.x86_64.rpm
		manageiq-appliance-tools-11.0.0-0.1.20200807153703.el8.x86_64.rpm  
		manageiq-gemset-11.0.0-0.1.20200807153703.el8.x86_64.rpm  
		manageiq-system-11.0.0-0.1.20200807153703.el8.x86_64.rpm
		[root@d50306c16c4b build_scripts]#cp /root/BUILD/rpms/x86_64 /output/
		[root@d50306c16c4b build_scripts]#
	```

3. To build manageiq-ui-worker clone the [manageiq-pods](https://github.com/ManageIQ/manageiq-pods). Copy all the rpms generated in step 2 to /mageiq-pods/images/manageiq-base/rpms, and do the container build as below:
	``` 
		git clone https://github.com/ManageIQ/manageiq-pods.git
		cd manageiq-pods/image/mnt/manageiq-pods/
		cp /mnt/manageiq-rpm_build/output/*.rpm .
		./bin/build -l -d images -r manageiq
		[root@localhost manageiq-pods]# podman images


		REPOSITORY                                     TAG      IMAGE ID       CREATED             SIZE
		localhost/manageiq/manageiq-ui-worker          latest   a25dffaa1015   13 hours ago        2.37 GB
	```
4. To run the manageiq container image 
	```
		[root@localhost myfork]# cd manageiq
		[root@localhost manageiq]# podman build -t manageiq/manageiq .
		[root@localhost manageiq]# podman run -di -p 8443:443 manageiq/manageiq
	```
        


		 

				 
