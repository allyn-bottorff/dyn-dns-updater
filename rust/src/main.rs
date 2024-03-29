use std::net::Ipv4Addr;
use std::str::FromStr;
use std::{error, fs};
use std::{thread, time};

use reqwest;
use serde::{Deserialize, Serialize};
use serde_json;

#[derive(Deserialize, Serialize)]
struct UnifiCreds {
    username: String,
    password: String,
}

#[derive(Deserialize)]
struct Config {
    unifi_creds: UnifiCreds,
    unifi_url: String,
    disable_unifi_tls_validation: bool,
    cftoken: String,
    watch_records: Vec<String>,
    poll_seconds: u64,
}

#[derive(Deserialize)]
struct UnifiHealth {
    data: Vec<SubSystemHealth>,
}

#[derive(Deserialize)]
struct SubSystemHealth {
    subsystem: String,
    wan_ip: Option<String>,
}

#[derive(Deserialize, Debug)]
struct Zone {
    name: String,
    id: String,
}

#[derive(Deserialize)]
struct ZoneResult {
    result: Vec<Zone>,
}

#[derive(Deserialize)]
struct ZoneDNSResponse {
    // success: bool,
    result: Vec<Record>,
}

#[derive(Deserialize)]
struct Record {
    id: String,
    r#type: String,
    name: String,
    content: String,
}

type Result<T> = std::result::Result<T, Box<dyn error::Error>>;

/// Read secrets and config
/// Get IP from Unifi
/// Get IP from Cloudflare
/// Check if different
/// -- If different, update zone apex records in Cloudflare
/// Sleep 10 minutes or so.
fn main() {

    let conf = match read_config() {
        Ok(c) => c,
        Err(e) => {
            println!("Failed to read config: {}", e);
            panic!();
            // continue;
        }
    };
    let mut u_client = reqwest::blocking::Client::builder()
        .cookie_store(true)
        .build()
        .unwrap();

    if conf.disable_unifi_tls_validation {
    u_client = reqwest::blocking::Client::builder()
        .danger_accept_invalid_certs(true)
        .cookie_store(true)
        .build()
        .unwrap();
    }


    // Loop forever with a thread sleep
    loop {
        thread::sleep(time::Duration::new(conf.poll_seconds, 0));

        println!("Starting IP Sync...");
        let body = match serde_json::to_string(&conf.unifi_creds) {
            Ok(b) => b,
            Err(e) => {
                println!("Failed to parse config: {}", e);
                panic!();
                // continue;
            }
        };
        let resp = match u_client
            .post(format!("{}/api/login", &conf.unifi_url))
            .body(body)
            .send()
        {
            Ok(r) => r,
            Err(e) => {
                println!("Failed to login to UniFi controller: {}", e);
                // panic!();
                continue;
            }
        };

        if resp.status() != reqwest::StatusCode::OK {
            println!("UniFi login failed: {}", resp.status().to_string());
            // panic!();
            continue;
        }

        println!("Unifi login successsful");

        let wanip = match get_unifi_ip(&mut u_client, &conf) {
            Ok(opt) => match opt {
                Some(ip) => ip,
                None => {
                    println!("Failed to parse IP from unifi");
                    // panic!();
                    continue;
                }
            },
            Err(e) => {
                println!("Failed to retrieve WAN IP: {}", e);
                // panic!();
                continue;
            }
        };

        println!("WAN IP: {}", wanip.to_string());

        let c_client = reqwest::blocking::Client::new();

        let zones = match get_zones(&c_client, &conf.cftoken) {
            Ok(z) => z,
            Err(e) => {
                println!("Failed to get zones from CloudFlare: {}", e);
                // panic!();
                continue;
            }
        };

        // println!("Cloudflare Zones:");
        // println!("{:?}", zones);

        for zone in zones {
            if conf.watch_records.contains(&zone.name) {
                println!("Checking zone: {}", zone.name);
                let apex = match get_apex(&zone, &c_client, &conf.cftoken) {
                    Ok(opt) => match opt {
                        Some(apex) => apex,
                        None => {
                            println!("Failed to get apex A record for zone: {}", zone.name);
                            // panic!();
                            continue;
                        }
                    },
                    Err(e) => {
                        println!("Failed to get apex A record: {}", e);
                        // panic!();
                        continue;
                    }
                };
                println!("Zone apex: {}", apex.content);
                if apex.content != wanip.to_string() {
                    println!("Found mismatched zone: {}", zone.name);
                    match update_apex(&zone, &apex, &wanip, &c_client, &conf.cftoken) {
                        Ok(()) => {
                            println!("Updated {} with IP: {}", &zone.name, wanip.to_string());
                        }
                        Err(e) => {
                            println!("Failed to update IP: {}", e);
                            // panic!();
                            continue;
                        }
                    };
                    //update record
                }
            }
        }
    }
}

/// Read Unifi login credentials from a file
fn read_config() -> Result<Config> {
    let contents = fs::read_to_string("/secrets/config.json")?;

    let creds: Config = serde_json::from_str(&contents)?;

    Ok(creds)
}

/// Get the IP address of the WAN interface from the Unifi Controller
fn get_unifi_ip(client: &mut reqwest::blocking::Client, conf: &Config) -> Result<Option<Ipv4Addr>> {
    let resp = client
        .get(format!("{}/api/s/default/stat/health", &conf.unifi_url))
        .send()?
        .error_for_status()?;

    let health: UnifiHealth = resp.json()?;

    for subsystem in health.data {
        if subsystem.subsystem == "wan" {
            let ipstr = match subsystem.wan_ip {
                Some(ip) => ip,
                None => {
                    continue;
                }
            };
            let wanip = Ipv4Addr::from_str(&ipstr)?;
            return Ok(Some(wanip));
        }
    }
    Ok(None)
}

/// Get the zones from CloudFlare
fn get_zones(client: &reqwest::blocking::Client, token: &String) -> Result<Vec<Zone>> {
    let resp = client
        .get("https://api.cloudflare.com/client/v4/zones")
        .header("Authorization", format!("Bearer {}", token))
        .send()?
        .error_for_status()?;

    let zone_result: ZoneResult = resp.json()?;

    Ok(zone_result.result)
}

/// Get the apex A record from a zone
fn get_apex(
    zone: &Zone,
    client: &reqwest::blocking::Client,
    token: &String,
) -> Result<Option<Record>> {
    let resp = client
        .get(format!(
            "https://api.cloudflare.com/client/v4/zones/{}/dns_records",
            zone.id
        ))
        .header("Authorization", format!("Bearer {}", token))
        .send()?
        .error_for_status()?;

    let z_result: ZoneDNSResponse = resp.json()?;

    for rec in z_result.result {
        if rec.name == zone.name {
            if rec.r#type == "A" {
                return Ok(Some(rec));
            }
        }
    }
    Ok(None)
}

/// Update an A record
fn update_apex(
    zone: &Zone,
    record: &Record,
    wanip: &Ipv4Addr,
    client: &reqwest::blocking::Client,
    token: &String,
) -> Result<()> {
    let body = format!("{{\"content\": \"{}\"}}", wanip.to_string());
    let _resp = client
        .patch(format!(
            "https://api.cloudflare.com/client/v4/zones/{}/dns_records/{}",
            zone.id, record.id
        ))
        .header("Authorization", format!("Bearer {}", token))
        .header("Content-Type", "application/json")
        .body(body)
        .send()?
        .error_for_status()?;

    Ok(())
}
