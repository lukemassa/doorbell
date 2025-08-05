
use std::thread;
use std::time::Duration;

static BASE_HEALTH_URL: &str = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50";

static UPDATE_FREQ: &Duration = &Duration::from_secs(10*60);


fn update_healthcheck(status: bool) {

    let url = if status {
        BASE_HEALTH_URL.to_string()
    } else {
        format!("{}/fail", BASE_HEALTH_URL)
    };

    let client = reqwest::blocking::Client::new();
    let res = client.get(&url).send();

    match res {
        Ok(response) => match response.text() {
            Ok(body) => println!("Posted to {}: {}", url, body),
            Err(e) => eprintln!("Failed to read body from {}: {}", url, e),
        },
        Err(e) => eprintln!("Failed to post to {}: {}", url, e),
    }
}

fn main() {
    loop {
       update_healthcheck(true);
        thread::sleep(*UPDATE_FREQ);
    }
}
