extern crate hyper;
extern crate rustbox;
extern crate rustc_serialize;

#[allow(dead_code)]

use std::io::Read;
use std::default::Default;
use std::process::Command;
use std::process::Stdio;

use hyper::Client;
use hyper::header::Connection;

use rustbox::{RustBox, Color, Key};
use rustc_serialize::json;
// use rustc_serialize::json::Json;

const NEW: &'static str = "https://hacker-news.firebaseio.com/v0/topstories.json";
const BASE: &'static str = "https://hacker-news.firebaseio.com/v0/item";
const MAX: usize = 35;

struct Style {
    style: rustbox::Style,
    fg: Color,
    bg: Color,
}

#[derive(RustcDecodable, Debug)]
struct Item {
    title: String,
    id: u64,
    url: Option<String>,
}

//impl Item {
    //fn show(&self) -> String {
        //format!("{}-{}", self.title, self.descendants)
    //}
//}


fn get_content(client: &Client, url: &str) -> String {
    let mut res = client.get(url)
                        .header(Connection::keep_alive())
                        .send()
                        .unwrap();
    let mut body = String::new();
    res.read_to_string(&mut body).unwrap();

    body
}

fn main() {

    let mut stories = Vec::new();

    let normal = Style {
        style: rustbox::RB_NORMAL,
        fg: Color::Default,
        bg: Color::Black,
    };
    let focused = Style {
        style: rustbox::RB_BOLD,
        fg: Color::Green,
        bg: Color::Black,
    };

    let rustbox = match RustBox::init(Default::default()) {
        Result::Ok(v) => v,
        Result::Err(e) => panic!("{}", e),
    };

    let mut row: usize = 0;
    let client = Client::new();
    let res: String = get_content(&client, NEW);
    let tops: Vec<usize> = json::decode(res.as_str()).unwrap();
    let mut i: usize = 0;
    for id in tops {
        if i > MAX {
            break;
        }
        let mut url: String = String::from(BASE);
        url.push_str("/");
        url.push_str(id.to_string().as_str());
        url.push_str(".json");
        let res: String = get_content(&client, url.as_str());
        let item: Item = json::decode(res.as_str()).unwrap();
        if i == 0 {
            rustbox.print(1,
                          i,
                          focused.style,
                          focused.fg,
                          focused.bg,
                          item.title.as_str());
        } else {
            rustbox.print(1,
                          i,
                          normal.style,
                          normal.fg,
                          normal.bg,
                          item.title.as_str());
        }
        i = i + 1;
        stories.push(item);
    }

    rustbox.present();

    loop {
        match rustbox.poll_event(false) {
            Ok(rustbox::Event::KeyEvent(key)) => {
                match key {
                    Key::Char('q') => break,
                    Key::Char('j') => {
                        if row < MAX {
                            rustbox.print(1,
                                          row,
                                          normal.style,
                                          normal.fg,
                                          normal.bg,
                                          stories[row].title.as_str());
                            row = row + 1;
                            rustbox.print(1,
                                          row,
                                          focused.style,
                                          focused.fg,
                                          focused.bg,
                                          stories[row].title.as_str());
                            rustbox.present();
                        }
                    }
                    Key::Char('k') => {
                        if row > 0 {
                            rustbox.print(1,
                                          row,
                                          normal.style,
                                          normal.fg,
                                          normal.bg,
                                          stories[row].title.as_str());
                            row = row - 1;
                            rustbox.print(1,
                                          row,
                                          focused.style,
                                          focused.fg,
                                          focused.bg,
                                          stories[row].title.as_str());
                            rustbox.present();
                        }
                    }
                    Key::Char('o') => {
                        match stories[row].url {
                            Some(ref d) => {
                                Command::new("xdg-open")
                                    .arg(d.clone().as_str())
                                    .stdout(Stdio::null())
                                    .stderr(Stdio::null())
                                    .spawn()
                                    .unwrap();
                                ()
                            }
                            None => (),
                        }
                    }

                    Key::Char('c') => {
                        Command::new("xdg-open")
                            .arg(format!("https://news.ycombinator.com/item?id={}", stories[row].id).as_str())
                            .stdout(Stdio::null())
                            .stderr(Stdio::null())
                            .spawn()
                            .unwrap();
                    }
                    _ => (),
                }
            }
            Err(e) => panic!("{}", e),
            _ => {}
        }
    }
}
