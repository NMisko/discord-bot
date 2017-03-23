# discord-bot

A **bot** for my discord server. Plays **music**, predicts the **weather**, throws **coins** and **dices**, **reminds** people of stuff, tells LoL **elo**, starts **polls** and **chats**.

Should theoretically be scalable to multiple servers, this was never tested though.

`/cmd/bot/` contains the bot, which requires the following flags to run:

  - -t \<discord authentication token\>
  - -k \<riot API key\>
  - -c \<cleverbot API key\>
  - -o \<discord ID of the owner\>

`/cmd/botsettings/` can be used to edit a bots appearance and requires some the following flags:
  - -t \<discord authentication token\>
  - -n \<new nickname\>
  - -f \<new profile image\>
  - -s \<new bot status\>
