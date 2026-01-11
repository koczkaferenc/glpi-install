# Glpi-install

A projekt célja egy GLPI és kapcsolódó adatbázisa, egy nginx proxy telepítése és tanúsítvány generálása. A szerver telepítése után a GLPI alap beállítása, Windows és Linux kliensek bekapcsolása.

Paraméterek, melyeket minden fájlban és az útmutatóban is a megfelelő domain névvel kell helyettesíteni:

* CTID: A konténer ID-je
* IP: a konténer IP címe <IP>
* Gw: A konténer átjárója <GW>
* Hostname: gl.example.hu

## CT készítése

* Image: debian-13-standard_13.1-2_amd64
* VM: Debian
* CTID: 263
* Hostname: docker
* Disk: 64G
* Cores: 4
* RAM: 8192/2048
* IP: <IP>/24
* Gw: <GW>

## Konténer beállítása
```shell
apt update
apt -y upgrade
apt -y install mc vim ssh docker.io docker-compose certbot at
sed -i 's/#Port 22/Port 65061/' /etc/ssh/sshd_config 
sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config 
systemctl restart ssh

mkdir -p /var/docker/glpi.example.hu
cd /var/docker/glpi.example.hu
```

## Certek generlása

A gl.DOMAIN.hu címnek a szervezet külső IP címére kell mutatnia, ez feltétlenül szükséges az érvényes Letsencrypt tanúsítvány létehozásához és a későbbi frissítésekhez.

Egy cert generálása:

```bash
certbot certonly --standalone -d gl.example.hu
```

Ez után a konténer indítása a ```run``` paranccsal történik.

## Szerver beállításai

1. Login: glpi/glpi
1. Dashboard: Disable demonstration
1. Admin user átnevezés, jelszócsere, majd: post-only, 
1. Administration/Inventory/Enable Inventory: a kliensek automatikus felvételének engedélyezése.
1. Helyek felvétele
1. Nyomtatók felvétele
1. Hálózati eszközök felvétele
1. UPS-ek felvétele
1. SIM kártyák felvétele

## Nézet beállítások:

Antivirus, sorozatszám.

## Kliensek:

Linux: 
```bash
apt remove --purge wazuh-agent wazuh-manager
rm /etc/apt/sources.list.d/wazuh.list

cd /var/tmp
wget https://github.com/glpi-project/glpi-agent/releases/download/1.13/glpi-agent-1.13-linux-installer.pl
perl glpi-agent-1.13-linux-installer.pl --server="https://gl.example.hu/front/inventory.php" --runnow --install
```

Restart/Státusz:
```bash
systemctl status glpi-agent
systemctl restart glpi-agent
```

Windows:
A telepítést powershellel végezzük:

```bash
winget install glpi-agent --custom="SERVER='http://gl.example.hu/front/inventory.php' RUNNOW=1"
```
