#!/bin/bash

# Vérifie si Node.js est installé
if ! command -v node &> /dev/null
then
    echo "Node.js n'est pas installé. Veuillez l'installer pour exécuter ce script."
    exit 1
fi

# Chemin vers le fichier iutimemanager.js (ajustez-le selon votre structure de répertoire)
chemin_fichier="iutimemanager.js"

# Vérifie si le fichier existe
if [ ! -f "$chemin_fichier" ]
then
    echo "Le fichier iutimemanager.js n'existe pas dans le chemin spécifié : $chemin_fichier"
    exit 1
fi

# Exécute le fichier avec Node.js
node "$chemin_fichier"

