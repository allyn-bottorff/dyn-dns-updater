use std::net::Ipv4Addr;
use std::str::FromStr;
use std::{error, fs};
// use std::{thread, time};

use reqwest;
use serde::{Deserialize, Serialize};
use serde_json;

#[derive(Deserialize, Serialize)]
struct UnifiCreds {
    username: String,
    password: String,
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

#[derive(Deserialize)]
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
    success: bool,
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
    println!("Starting IP Sync...");

    // loop {

    let mut u_client = reqwest::blocking::Client::builder()
        .cookie_store(true)
        .build()
        .unwrap();

    let u_creds = match read_secrets() {
        Ok(c) => c,
        Err(e) => {
            println!("{}", e);
            panic!();
            // continue;
        }
    };
    let body = match serde_json::to_string(&u_creds) {
        Ok(b) => b,
        Err(e) => {
            println!("{}", e);
            panic!();
            // continue;
        }
    };
    let resp = match u_client
        .post("https://unifi.b6f.net/api/login")
        .body(body)
        .send()
    {
        Ok(r) => r,
        Err(e) => {
            println!("{}", e);
            panic!();
            // continue;
        }
    };

    if resp.status() != reqwest::StatusCode::OK {
        println!("UniFi login failed: {}", resp.status().to_string());
        panic!();
        // continue;
    }

    println!("Unifi login successsful");

    let wanip = match get_unifi_ip(&mut u_client) {
        Ok(opt) => match opt {
            Some(ip) => ip,
            None => {
                println!("Failed to parse IP from unifi");
                panic!();
                // continue;
            }
        },
        Err(e) => {
            println!("{}", e);
            panic!();
            // continue;
        }
    };

    println!("WAN IP: {}", wanip.to_string());

    // thread::sleep(time::Duration::new(30, 0));
    // }
}

/// Read Unifi login credentials from a file
fn read_secrets() -> Result<UnifiCreds> {
    let contents = fs::read_to_string("./creds.json")?;

    let creds: UnifiCreds = serde_json::from_str(&contents)?;

    Ok(creds)
}

/// Get the IP address of the WAN interface from the Unifi Controller
fn get_unifi_ip(client: &mut reqwest::blocking::Client) -> Result<Option<Ipv4Addr>> {
    let resp = client
        .get("https://unifi.b6f.net/api/s/default/stat/health")
        .send()?;

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
fn get_zone(token: String) {}
