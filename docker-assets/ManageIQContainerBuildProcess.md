## Building and running a ManageIQ Container 

The following steps describe how to build a ManageIQ container which also includes private manageiq <span style="color:blue">provider</span> plugin.

1. In ManageIQ core, developer can override gem dependencies in directory bundler.d/local_plugins.rb. This small helper under bundler.d directory points to the developers version of provider plugin gem. For a given provider with private manageiq-<span style="color:blue">*providers*</span>-plugin gem, add the following lines in bundler.d/local_plugins.rb.

	```
	group :plugin_group_name, :manageiq_default do
  	  gem 'manageiq-providers-plugin',  :git =>  'https://manageiq/manageiq-providers-plugin.git',  :branch => "master"
	end
	```


	Developer may configure bundler  with git user login and password, or personal access token to download and install **manageiq-provider-plugin** from git during **bundle install**
	```
    bundle config GITHUB__COM username:password USER:XXXXXXXX --global
    bundle confg
     Settings are listed in order of priority. The top value will be used.
     github.com
     Set for the current user (/root/.bundle/config): "USER:XXXXXXXX"
   ```
	
   > *NOTE:* For an internal repo, developer can either do https/ssh credential setup or generate a personal access token or use username and password for cloning the private gem. The [personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) with full repo access can be setup in github.com page as:
	> <br>```settings > developer settings > personal access tokens > generate new token```</br>
	
	This will allow the bundler to fetch and install user plugin when doing ***bundle install*** or ***bin/setup***. 
	Add, commit and push these to your git fork.

2.	Create manageiq rpm
	- Clone the public **[manageiq-rpm_build](https://github.com/ManageIQ/manageiq-rpm_build)** repo
		> git clone https://github.com/ManageIQ/manageiq-rpm_build.git 

	- All build configuration changes can be made in `OPTIONS/options.yml` which can be volume mounted when running manageiq-rpm_build container. For example
		```
		---
        product_name: manageiq
        repos:
        ref:        master
          manageiq:
            url:      https://XXXXXXXX@github.com/ManageIQ/manageiq.git
            ref:      master
		``` 
      Here url points to developer local manageiq gem with XXXXXXXX as the GIT personal access token for this developer. This will download a copy of manageiq.git from the github.com during rpm build process
  
    - Create another local directory OUTPUT which can also be mounted to /root/BUILD directory to get access to the manageiq rpms 

	- Build the docker container.

	 	```[manageiq-rpm_build]$ docker build --tag manageiq-rpm:latest .```

	- Run the *manageiq-rpm:latest* container image, mounting OPTION and OUTPUT directories<br>
		```
		$ docker run -e CI_USER_TOKEN=$CI_USER_TOKEN -v $PWD/OPTIONS:/root/OPTIONS -v $PWD/OUTPUT:/root/BUILD -it manageiq-rpm:latest build
		~~~
		---> ManageIQ::RPMBuild::SetupSourceRepos#initialize
		~~~
		---> git clone --depth 1 -b master https://XXXXXXXX@github.com/USER/manageiq.git manageiq
		Cloning into 'manageiq'...
		remote: Enumerating objects: 3243, done.
		remote: Counting objects: 100% (3243/3243), done.
		~~~
		```

	- Once the RPMs are build we can copy these to mounted volume  *output* 
		```
		$ ls BUILD/rpms/x86_64
		manageiq-appliance-11.0.0-0.1.20200807153703.el8.x86_64.rpm        
		manageiq-core-11.0.0-0.1.20200807153703.el8.x86_64.rpm   
		manageiq-pods-11.0.0-0.1.20200807153703.el8.x86_64.rpm    
		manageiq-ui-11.0.0-0.1.20200807153703.el8.x86_64.rpm
		manageiq-appliance-tools-11.0.0-0.1.20200807153703.el8.x86_64.rpm  
		manageiq-gemset-11.0.0-0.1.20200807153703.el8.x86_64.rpm  
		manageiq-system-11.0.0-0.1.20200807153703.el8.x86_64.rpm
		```

3. To build manageiq-ui-worker clone the [manageiq-pods](https://github.com/ManageIQ/manageiq-pods). Copy all the rpms generated in step 2 to ```manageiq-pods/images/manageiq-base/rpms```, and do the container image build as below:
	``` 
	git clone https://github.com/ManageIQ/manageiq-pods.git
    cp manageiq-rpm_build/output/*.rpm manageiq-pods/image/mnt/manageiq-pods/
    cd manageiq-rpm_build
	./bin/build -l -d images -r manageiq
	docker images
	REPOSITORY                                     TAG      IMAGE ID       CREATED             SIZE
	localhost/manageiq/manageiq-ui-worker          latest   a25dffaa1015   13 hours ago        2.37 GB
	```
4. Now we have manageiq-ui-worker container in our local docker repo and we can use this to build manageiq/manageiq:latest docker container. Go to
	```
	$ cd manageiq
	$ docker build -t manageiq/manageiq .
	$ docker run -di -p 8443:443 manageiq/manageiq
	```
        


		 

				 
