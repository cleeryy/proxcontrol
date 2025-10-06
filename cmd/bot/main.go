package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
    "github.com/joho/godotenv"
    "github.com/cleeryy/ProxControl/internal/proxmox"
)

var (
    proxmoxClient *proxmox.Client
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("Aucun fichier .env trouvé")
    }

    token := os.Getenv("DISCORD_TOKEN")
    if token == "" {
        log.Fatal("DISCORD_TOKEN n'est pas défini")
    }

    allowedVMsStr := os.Getenv("ALLOWED_VMS")
    allowedVMs, err := proxmox.ParseAllowedVMs(allowedVMsStr)
    if err != nil {
        log.Fatal("Erreur lors du parsing des VMs autorisées:", err)
    }

    proxmoxClient = proxmox.NewClient(
        os.Getenv("PROXMOX_URL"),
        os.Getenv("PROXMOX_TOKEN_ID"),
        os.Getenv("PROXMOX_SECRET"),
        os.Getenv("PROXMOX_NODE"),
        allowedVMs,
    )

    log.Printf("VMs autorisées: %v", allowedVMs)

    dg, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal("Erreur lors de la création de la session Discord:", err)
    }

    dg.AddHandler(ready)
    dg.AddHandler(interactionCreate)

    dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

    err = dg.Open()
    if err != nil {
        log.Fatal("Erreur lors de l'ouverture de la connexion:", err)
    }
    defer dg.Close()

    fmt.Println("Bot démarré. Appuyez sur CTRL+C pour quitter.")

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
    log.Printf("Bot connecté en tant que %s#%s", event.User.Username, event.User.Discriminator)

    guildID := os.Getenv("DISCORD_GUILD_ID")

    commands := []*discordgo.ApplicationCommand{
        {
            Name:        "vm",
            Description: "Gérer les machines virtuelles Proxmox",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Name:        "start",
                    Description: "Démarrer une VM",
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:         discordgo.ApplicationCommandOptionString,
                            Name:         "vm",
                            Description:  "Nom ou ID de la VM à démarrer",
                            Required:     true,
                            Autocomplete: true, // Active l'autocomplete
                        },
                    },
                },
                {
                    Name:        "stop",
                    Description: "Arrêter une VM (graceful)",
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:         discordgo.ApplicationCommandOptionString,
                            Name:         "vm",
                            Description:  "Nom ou ID de la VM à arrêter",
                            Required:     true,
                            Autocomplete: true,
                        },
                    },
                },
                {
                    Name:        "status",
                    Description: "Voir le statut d'une VM",
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:         discordgo.ApplicationCommandOptionString,
                            Name:         "vm",
                            Description:  "Nom ou ID de la VM",
                            Required:     true,
                            Autocomplete: true,
                        },
                    },
                },
                {
                    Name:        "list",
                    Description: "Lister toutes les VMs autorisées",
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                },
            },
        },
    }

    for _, cmd := range commands {
        _, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
        if err != nil {
            log.Printf("Erreur lors de la création de la commande %s: %v", cmd.Name, err)
        } else {
            log.Printf("Commande %s enregistrée avec succès", cmd.Name)
        }
    }
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
    switch i.Type {
    case discordgo.InteractionApplicationCommand:
        handleCommand(s, i)
    case discordgo.InteractionApplicationCommandAutocomplete:
        handleAutocomplete(s, i)
    }
}

// handleAutocomplete gère les suggestions d'autocomplete
func handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
    data := i.ApplicationCommandData()

    if data.Name != "vm" {
        return
    }

    // Récupérer la valeur actuelle tapée par l'utilisateur
    var focusedValue string
    if len(data.Options) > 0 && len(data.Options[0].Options) > 0 {
        focusedValue = data.Options[0].Options[0].StringValue()
    }

    // Récupérer la liste des VMs
    vms, err := proxmoxClient.ListVMs()
    if err != nil {
        log.Printf("Erreur lors de la récupération des VMs: %v", err)
        return
    }

    // Créer les choix pour l'autocomplete
    choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
    focusedLower := strings.ToLower(focusedValue)

    for _, vm := range vms {
        // Filtrer par ce que l'utilisateur a tapé
        if focusedValue == "" || 
           strings.Contains(strings.ToLower(vm.Name), focusedLower) ||
           strings.Contains(fmt.Sprintf("%d", vm.VMID), focusedValue) {
            
            choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
                Name:  fmt.Sprintf("%s (ID: %d) - %s", vm.Name, vm.VMID, vm.Status),
                Value: fmt.Sprintf("%d", vm.VMID), // On retourne l'ID en string
            })

            // Discord limite à 25 choix max
            if len(choices) >= 25 {
                break
            }
        }
    }

    // Envoyer les suggestions
    err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionApplicationCommandAutocompleteResult,
        Data: &discordgo.InteractionResponseData{
            Choices: choices,
        },
    })

    if err != nil {
        log.Printf("Erreur lors de l'envoi de l'autocomplete: %v", err)
    }
}

// handleCommand gère l'exécution des commandes
func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    data := i.ApplicationCommandData()

    if data.Name != "vm" {
        return
    }

    options := data.Options
    if len(options) == 0 {
        return
    }

    subCommand := options[0]

    // Répondre immédiatement
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })

    var response string
    var color int

    switch subCommand.Name {
    case "list":
        vms, err := proxmoxClient.ListVMs()
        if err != nil {
            response = fmt.Sprintf("❌ Erreur lors de la récupération des VMs: %v", err)
            color = 0xFF0000
        } else {
            if len(vms) == 0 {
                response = "Aucune VM autorisée trouvée."
            } else {
                response = "**VMs autorisées:**\n\n"
                for _, vm := range vms {
                    statusEmoji := "🔴"
                    if vm.Status == "running" {
                        statusEmoji = "🟢"
                    }
                    response += fmt.Sprintf("%s **%s** (ID: %d) - %s\n", statusEmoji, vm.Name, vm.VMID, vm.Status)
                }
            }
            color = 0x0099FF
        }

    default:
        // Pour start, stop, status - récupérer le VMID
        if len(subCommand.Options) == 0 {
            return
        }

        vmInput := subCommand.Options[0].StringValue()
        vmid, err := parseVMInput(vmInput)
        if err != nil {
            response = fmt.Sprintf("❌ %v", err)
            color = 0xFF0000
        } else {
            response, color = executeVMCommand(subCommand.Name, vmid)
        }
    }

    // Envoyer la réponse finale
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Embeds: &[]*discordgo.MessageEmbed{
            {
                Description: response,
                Color:       color,
            },
        },
    })
}

// parseVMInput convertit l'input (nom ou ID) en VMID
func parseVMInput(input string) (int, error) {
    // Essayer de parser comme un entier d'abord
    var vmid int
    _, err := fmt.Sscanf(input, "%d", &vmid)
    if err == nil {
        return vmid, nil
    }

    // Sinon, chercher par nom
    vm, err := proxmoxClient.FindVMByName(input)
    if err != nil {
        return 0, err
    }

    return vm.VMID, nil
}

// executeVMCommand exécute une commande sur une VM
func executeVMCommand(command string, vmid int) (string, int) {
    switch command {
    case "start":
        err := proxmoxClient.StartVM(vmid)
        if err != nil {
            return fmt.Sprintf("❌ Erreur lors du démarrage de la VM %d: %v", vmid, err), 0xFF0000
        }
        return fmt.Sprintf("✅ VM %d en cours de démarrage", vmid), 0x00FF00

    case "stop":
        err := proxmoxClient.ShutdownVM(vmid)
        if err != nil {
            return fmt.Sprintf("❌ Erreur lors de l'arrêt de la VM %d: %v", vmid, err), 0xFF0000
        }
        return fmt.Sprintf("🛑 VM %d en cours d'arrêt", vmid), 0xFFA500

    case "status":
        status, err := proxmoxClient.GetVMStatus(vmid)
        if err != nil {
            return fmt.Sprintf("❌ Erreur lors de la récupération du statut de la VM %d: %v", vmid, err), 0xFF0000
        }

        uptimeHours := status.Uptime / 3600
        memUsedGB := float64(status.Mem) / 1024 / 1024 / 1024
        memTotalGB := float64(status.MaxMem) / 1024 / 1024 / 1024
        cpuPercent := status.CPU * 100

        response := fmt.Sprintf(
            "📊 **VM %d - %s**\n"+
                "• État: **%s**\n"+
                "• Uptime: %d heures\n"+
                "• CPU: %.1f%%\n"+
                "• RAM: %.2f GB / %.2f GB",
            status.VMID, status.Name, status.Status,
            uptimeHours, cpuPercent, memUsedGB, memTotalGB,
        )

        color := 0xFF0000
        if status.Status == "running" {
            color = 0x00FF00
        }
        return response, color
    }

    return "❌ Commande inconnue", 0xFF0000
}

