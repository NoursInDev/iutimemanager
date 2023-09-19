const fs = require('fs');
const Discord = require('discord.js');
const cron = require('node-cron');
const config = require('./config.json');

const client = new Discord.Client({ intents: [config.discord.intents] });

client.once('ready', () => {
  console.log(`Logged in as ${client.user.tag}`);
  // Obtenir la liste des IDs de serveur à partir de config.json
  const serverIds = Object.keys(config.channels_id);

  // Planifier la tâche pour chaque ID de serveur
  serverIds.forEach((serverId) => {
    cron.schedule('0 1 * * *', () => {
      executeGoCodeAndSendImage(serverId);
    });
  });
});

client.login(config.discord.token);

function getNearestMondayBehind() {
  const currentDate = new Date();
  const currentDayOfWeek = currentDate.getDay(); // 0 pour dimanche, 1 pour lundi, ..., 6 pour samedi

  const daysUntilMonday = currentDayOfWeek === 0 ? 6 : currentDayOfWeek - 1; // Nombre de jours jusqu'à lundi
  const mondayDate = new Date(currentDate);
  mondayDate.setDate(currentDate.getDate() - daysUntilMonday);

  // Format de la date en YYYYMMDD
  const year = mondayDate.getFullYear();
  const month = (mondayDate.getMonth() + 1).toString().padStart(2, '0');
  const day = mondayDate.getDate().toString().padStart(2, '0');
  return `${year}${month}${day}`;
}

function getNearestSundayAhead() {
  const currentDate = new Date();
  const currentDayOfWeek = currentDate.getDay(); // 0 pour dimanche, 1 pour lundi, ..., 6 pour samedi

  const daysUntilSunday = 7 - currentDayOfWeek; // Nombre de jours jusqu'à dimanche
  const sundayDate = new Date(currentDate);
  sundayDate.setDate(currentDate.getDate() + daysUntilSunday);

  // Format de la date en YYYYMMDD
  const year = sundayDate.getFullYear();
  const month = (sundayDate.getMonth() + 1).toString().padStart(2, '0');
  const day = sundayDate.getDate().toString().padStart(2, '0');
  return `${year}${month}${day}`;
}


function executeGoCodeAndSendImage(serverId) {
  const currentDate = new Date();
  const year = currentDate.getFullYear();
  const month = (currentDate.getMonth() + 1).toString().padStart(2, '0');
  const day = currentDate.getDate().toString().padStart(2, '0');
  const imageName = `calendar-${year}-${month}-${day}.png`;

  // Obtenez l'URL associée à l'ID du serveur à partir de config.json
  const url = config.planning[serverId];
  if (!url) {
    console.error(`L'URL n'a pas été trouvée pour le serveur ID ${serverId}.`);
    return;
  }

  // Obtenez la date de début et de fin de la semaine au format YYYYMMDD
// Utilisation :
  const startDate = getNearestMondayBehind();
  const endDate = getNearestSundayAhead();

  console.log(`StartDate: ${startDate}`);
  console.log(`EndDate: ${endDate}`);

  const { exec } = require('child_process');
  // Exécutez votre programme Go avec les paramètres appropriés (serverId, startDate, endDate)
  exec(`main ${serverId} ${startDate} ${endDate}`, (error, stdout, stderr) => {
    if (error) {
      console.error(`Erreur lors de l'exécution du programme Go : ${error}`);
      return;
    }
    console.log(`Sortie du programme Go : ${stdout}`);
    sendImageToDiscord(serverId, imageName);
  });
}

function sendImageToDiscord(serverId, imageName) {
  const channelId = config.channels_id[serverId];
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
