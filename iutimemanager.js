const Discord = require('discord.js');
const fs = require('fs');
const config = require('./config.json');
const client = new Discord.Client({ 
  intents: [config.intents]
});

client.once('ready', () => {
  console.log(`logged in as ${client.user.tag}`);
});


// Connexion au serveur Discord en utilisant le token du fichier config.json
client.login(config.token);