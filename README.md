# GalaNodes-Monitor
![This is an image](https://blogfiles.pstatic.net/MjAyMjA0MTBfMTky/MDAxNjQ5NTYzMDk5MDU0.ASj7X79CEHSUbGZB7-RHwyvZBttN_dFy1LIRmgDFBkwg.AEbIxIBAslSNH_zN9Pj3_cgLHeD_MliVAVv67eBhr50g.PNG.wcgclan/architecture.png)<br />
This tool is to operate and manage multiple Gala Games(GALA) nodes including Founders, TownStar and Music nodes. It's focused on managing multiple nodes.
# Loadmap
### Stage 1
- [x] Add & remove nodes lists
- [x] Config monitoring and notification settings
- [x] Regular updating nodes stats
- [x] Notify error state
- [x] Remote commands(stats, restart nodes, reboot OS, etc)
### Stage 2
- [ ] Restart nodes or reboot OS when it meets specific conditions.
- [ ] Save encrypted username and password of nodes(currently saved in plain text)
- [ ] Collect nodes resources(CPU, memory, network bandwith, etc.)
### Stage 3
- [ ] Built-in Web UI to control nodes and check events history
- [ ] Discord Bot to control your nodes by customized discord commands

# Configuration (config/config.json)
**Settings**<br />
It's common settings on all nodes.
```
{
    "settings" : {
        "webhookurl": "your webhook url",
        "monitorinterval" : 90,
        "regularreportinterval" : 3600,
        "discordnotifysnooze" : 600,
        "errortolerance" : {
            "Count" : 3,
            "Command" : "sudo reboot"
        }
    },
 ```
- **_webhookurl_** : the webhook url which you want to notify nodes errors.
  ![This is an image](https://blogfiles.pstatic.net/MjAyMjA0MTBfNzQg/MDAxNjQ5NTU3NzMxMTY4.w2o2lCaF-E1nWAeG_m3f9hLctyJbCFN0HpwTxE8n-vQg.gfx1LesfsCZP2gmVO9s-FQVtPAsNkMhF21XeevOaAh0g.JPEG.wcgclan/discord_alert.JPG)<br />
  - https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks<br />
- **_monitorinterval_** : executes "gala-node stats" every **_monitorinterval_** seconds on registered servers. 
- **_regularreportinterval_** : notify nodes state every **_regularreportinterval_** seconds.\
![This is an image](https://blogfiles.pstatic.net/MjAyMjA0MTBfMTk5/MDAxNjQ5NTU3Nzg3NjYx.3xJ8PzpFzHC_D45d9M6OUdBjr1ioaGSCMNNCyOj9i-og.-9UXaO3jQY5YvLxsJWNc5nRSMrKZkXDIFpmSwJ1U_Xkg.PNG.wcgclan/NodeReport.png)
- **_discordnotifysnooze_** : snooze discord notification for **_discordnotifysnooze_** seconds once it's done.
- **_errortolerance_**<br />
  - **_Count_** : notify discord when an error occurs **_errortolerance_** times in a row.<br />
  - **_Command_** : run a custom command("_**sudo reboot**_") when an error occurs **_errortolerance_** times in a row.<br />

**Servers**<br />
 ```
{
    "servers" : [
        {
            "name" : "server name",
            "address" : "your server address",
            "port" : 22,
            "username" : "ubuntu",
            "password" : "your password",
            "privatekeypath" : "",
            "nodes" : ["founders", "townstar"]
        },
 ```
- **_name_** : alias
- **_address_** : server ip address
- **_port_** : SSH port
- **_username_** : user name for SSH
- **_privatekeypath_** : the path of a key pair, consisting of a public key and a private key, is a set of security credentials
  - https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
- **_nodes_** : nodes lists to monitor. If you run only TownStar node, set only **_"townstar"_**. Currently supporting founders and townstar nodes.
# Command lists
- **_help_** : prints available command lists.
- **_discord_** : discord webhook test.
- **_nodes_** : prints the current status of nodes.
- **_find [key]_** : find nodes which have a key in name, address, machine ID.
- **_save_** : output nodes information to output.csv.
- **_cls_** : clear screen
- **_stats_** : run "gala-node stats" on all nodes.
- **_restart [index|address]_** : run "sudo systemctl restart gala-node" on a gala node or all nodes. The index is non zero based in nodes commands.
- **_reboot [index|address|all]_** : reboot operating system of a node or all nodes. The index is non zero based. The index is non zero based in nodes commands.
- **_exit_** : exit program.
# License
GalaNodes Monitor is provided under the [MIT License](https://github.com/woodsshin/GalaNodes-Monitor/blob/main/LICENSE).
# Contact
- Discord : **_woodsshin#3643_**
- Email : **_woods.shin1014@gmail.com_**
# Donation
- You can find more tools here. Your support is helpful to operate background programs for these tools.
- https://www.buymeacoffee.com/alpaca007
![This is an image](https://cdn.buymeacoffee.com/uploads/project_updates/2022/04/7b1182aa7d3b5da8f943eed203468856.png@1200w_0e.webp)
