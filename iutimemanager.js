const fs = require('fs');
const Discord = require('discord.js');
const cron = require('node-cron');
const config = require('./config.json');

const client = new Discord.Client({intents:[config.discord.intents]});

client.once('ready', () => {
  console.log(`Logged in as ${client.user.tag}`);
  // Planning daily task at 1 a.m.
  cron.schedule('0 1 * * *', () => {
    executeGoCodeAndSendImage();
  });
});

client.login(config.discord.token);

function executeGoCodeAndSendImage() {
  const currentDate = new Date();
  const year = currentDate.getFullYear();
  const month = (currentDate.getMonth() + 1).toString().padStart(2, '0');
  const day = currentDate.getDate().toString().padStart(2, '0');
  const imageName = `calendar-${year}-${month}-${day}.png`;

   const { exec } = require('child_process');
   exec('generator', (error, stdout, stderr) => {
    if (error) {
       console.error(`Erreur lors de l'exécution du programme Go : ${error}`);
       return;
     }
     console.log(`Sortie du programme Go : ${stdout}`);
     sendImageToDiscord(imageName);
  });

  // Envoyer l'image dans le salon Discord
  sendImageToDiscord(imageName);
}

function sendImageToDiscord(imageName) {
  const channelId = config.discord.salon;
  const channel = client.channels.cache.get(channelId);

  if (!channel) {
    console.error(`Le salon avec l'ID ${channelId} n'a pas été trouvé.`);
    return;
  }

  const imagePath = `./calendars/${imageName}`;

  // Vérifier si le fichier image existe
  if (fs.existsSync(imagePath)) {
    const attachment = new Discord.MessageAttachment(imagePath);
    channel.send(`Image du calendrier pour aujourd'hui :`, attachment);
  } else {
    console.error(`Le fichier image ${imageName} n'a pas été trouvé.`);
  }
}
