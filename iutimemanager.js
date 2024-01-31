const fs = require('fs');
const Discord = require('discord.js');
const cron = require('node-cron');
const config = require('./config.json');

const client = new Discord.Client({ intents: [config.discord.intents] });

client.once('ready', () => {
  console.log(`Logged in as ${client.user.tag}`);
  // Obtain the list of server IDs from config.json
  const serverIds = Object.keys(config.channels_id);

  // Schedule the task for each server ID
  serverIds.forEach((serverId) => {
    cron.schedule('0 8 * * *', () => {
      executeGoCodeAndSendImage(serverId);
    });
  });
});

client.login(config.discord.token);

// Function to get the nearest Monday before the current date
function getNearestMondayBehind() {
  const currentDate = new Date();
  const currentDayOfWeek = currentDate.getDay(); // 0 for Sunday, 1 for Monday, ..., 6 for Saturday

  const daysUntilMonday = currentDayOfWeek === 0 ? 6 : currentDayOfWeek - 1; // Number of days until Monday
  const mondayDate = new Date(currentDate);
  mondayDate.setDate(currentDate.getDate() - daysUntilMonday);

  // Date format in YYYYMMDD
  const year = mondayDate.getFullYear();
  const month = (mondayDate.getMonth() + 1).toString().padStart(2, '0');
  const day = mondayDate.getDate().toString().padStart(2, '0');
  return `${year}${month}${day}`;
}

// Function to get the nearest Sunday ahead of the current date
function getNearestSundayAhead() {
  const currentDate = new Date();
  const currentDayOfWeek = currentDate.getDay(); // 0 for Sunday, 1 for Monday, ..., 6 for Saturday

  const daysUntilSunday = 7 - currentDayOfWeek; // Number of days until Sunday
  const sundayDate = new Date(currentDate);
  sundayDate.setDate(currentDate.getDate() + daysUntilSunday);

  // Date format in YYYYMMDD
  const year = sundayDate.getFullYear();
  const month = (sundayDate.getMonth() + 1).toString().padStart(2, '0');
  const day = sundayDate.getDate().toString().padStart(2, '0');
  return `${year}${month}${day}`;
}

// Function to execute Go code and send an image to Discord
function executeGoCodeAndSendImage(serverId) {
  const currentDate = new Date();
  const year = currentDate.getFullYear();
  const month = (currentDate.getMonth() + 1).toString().padStart(2, '0');
  const day = currentDate.getDate().toString().padStart(2, '0');
  const imageName = `calendar-${year}-${month}-${day}_${serverId}.png`;

  // Get the URL associated with the server ID from config.json
  const url = config.planning[serverId];
  if (!url) {
    console.error(`URL not found for server ID ${serverId}.`);
    return;
  }

  // Get the start and end date of the week in YYYYMMDD format
  const startDate = getNearestMondayBehind();
  const endDate = getNearestSundayAhead();

  console.log(`StartDate: ${startDate}`);
  console.log(`EndDate: ${endDate}`);

  const { exec } = require('child_process');
  // Execute your Go program with the appropriate parameters (serverId, startDate, endDate)
  process.env.PATH = `${process.env.PATH}:/home/tbran/Documents/DEV/iutimemanager`;
  exec(`main ${serverId} ${startDate} ${endDate}`, (error, stdout, stderr) => {
    if (error) {
      console.error(`Error when executing the Go program: ${error}`);
      return;
    }
    console.log(`Go program output: ${stdout}`);
    sendImageToDiscord(serverId, imageName);
  });
}

// Function to send an image to a Discord channel
function sendImageToDiscord(serverId, imageName) {
  const channelId = config.channels_id[serverId];
  const channel = client.channels.cache.get(channelId);

  if (!channel) {
    console.error(`Channel with ID ${channelId} not found.`);
    return;
  }

  const imagePath = `./calendars/${imageName}`;

  // Check if the image file exists
  if (fs.existsSync(imagePath)) {
    const file = new Discord.AttachmentBuilder(imagePath);

    const exampleEmbed = {
      title: 'EMPLOI DU TEMPS DU JOUR',
    };

    channel.send({ files: [file] });
    channel.send({ embeds: [exampleEmbed] });
  } else {
    console.error(`Image file ${imageName} not found.`);
  }
}
