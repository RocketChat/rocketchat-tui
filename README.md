# rocketchat-tui

GSOC project to create a TUI using the charm UI components and Rocket.Chat's go sdk

```
                                                                                
                                                                                
                                                                                
                                                                                
                                                                                
                                                                                
          #################                                                     
           ####################                                                 
             #######################################                            
               ###########################################&                     
                ##################            %###############                  
               ###########                            %##########               
             (#######&                                     ########             
            #######                                          ########           
           ######                                              #######          
          ######          #####      %####%      #####          ######          
          ######         #######    ########    #######         %#####%         
          ######         %#####      ######      #####          ######          
           ######                                              #######          
            ######                                            ######(           
              #######                                      ########             
                ######                                 %#########               
               &#####%        (&                &#############                  
               ######       #############################(                      
             #######    ###########################                             
           ###################(                                                 
          #################                                                     
                                                                                
                                                                                
                                                                                
                                                                                
                                                                                

```

## To test it follow the below steps
- Clone this Project Repo and I'm assuming that you have locally setuped RocketChat Meteor Application in your machine.
- Run Rocketchat Meteor Server on your `http://localhost:3000` and login/signup into a new account save your credentials for future.
- In the RocketChat TUI root folder run `go get` in terminal to get all golang packages we are using
- Make a `.env` file in the project root directory and add below enteries in it.

    ```
    EMAIL=YOUR_LOCALHOST_ROCKETCHAT_EMAIL
    PASS=YOUR_LOCALHOST_ROCKETCHAT_PASSWORD
    DEV_SERVER_URL=http://localhost:3000
    ```
- Now in the RocketChat TUI root folder run `go run *.go` to run the TUI
- Hopefully you will see the TUI running.

**Note** - Things may break because it's under development