const puppeteer = require('puppeteer');
const fs = require('fs');

//Start the browser and create a browser instance
let browserInstance = startBrowser();
scrapeAll(browserInstance);

async function startBrowser() {
    let browser;
    try {
        console.log("Opening the browser......");
        browser = await puppeteer.launch({
            headless: false,
            args: ["--disable-setuid-sandbox"],
            'ignoreHTTPSErrors': true
        });
    } catch (err) {
        console.log("Could not create a browser instance => : ", err);
    }
    return browser;
}

async function scrapeAll(browserInstance) {
    let browser;
    try {
        browser = await browserInstance;

        let scrapedData;

        // get all enclosures from RSS with `npx podcast-dl --url <PODCAST_RSS_URL>`

        // nihongo picnic
        // scrapedData = await picnicScraperObject.scraper(browser);
        // await browser.close();
        // fs.writeFile("nihongo-picnic-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
        //     if (err) {
        //         return console.log(err);
        //     }
        //     console.log("The data has been scraped and saved successfully! View it at './picnic-data.json'");
        // });

        // // sakura 
        // scrapedData = await sakuraScraperObject.scraper(browser);
        // await browser.close();
        // fs.writeFile("sakura-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
        //     if (err) {
        //         return console.log(err);
        //     }
        //     console.log("The data has been scraped and saved successfully! View it at './sakura-data.json'");
        // });

        // // noriko
        // scrapedData = await norikoScraperObject.scraper(browser);
        // await browser.close();
        // fs.writeFile("noriko-season1-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
        //     if (err) {
        //         return console.log(err);
        //     }
        //     console.log("The data has been scraped and saved successfully! View it at './noriko-data.json'");
        // });

        // // cj
        // scrapedData = await cjScraperObject.scraper(browser);
        // await browser.close();
        // fs.writeFile("cj-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
        //     if (err) {
        //         return console.log(err);
        //     }
        //     console.log("The data has been scraped and saved successfully! View it at './cj-data.json'");
        // });

    }
    catch (err) {
        console.log("Could not resolve the browser instance => ", err);
    }
}

const cjScraperObject = {
    url: 'https://cijapanese.com/',
    patreonUrl: 'https://cijapanese.com/dragon-ball-memories-members-only/',
    async scraper(browser) {
        let page = await browser.newPage();

        let audioURL = 'https://drive.google.com/drive/folders/1qfYzeDKplz-CWxIAzLcayuHOeAO3bIlt';

        console.log(`Navigating to ${this.url}...`);
        // Navigate to the selected page
        await page.goto(this.url, { waitUntil: 'networkidle0' });

        await page.click('.level-tabs-container div:last-child');



        // get links

        let hrefs = [];
        for (; ;) {
            let done = await page.evaluate(() => {
                let current = document.querySelectorAll('.pager-bar .current')[1];
                return current.nextElementSibling.classList.contains('disabled');
            });

            let links = await page.evaluate(() => {
                let links = [];

                // get links
                let as = document.querySelectorAll('.post-item-container')[1].querySelectorAll('a');
                for (a of as) {
                    //console.log(a.href);
                    links.push(a.href);
                }

                document.querySelectorAll('.pager-bar .current')[1].nextElementSibling.click();

                return links;
            });

            hrefs = hrefs.concat(links);

            if (done) {
                break;
            }
        }

        console.log('num stories: ', hrefs.length);

        // patreon login
        {
            console.log(`Patreon login ${this.patreonUrl}...`);
            await page.goto(this.patreonUrl, { waitUntil: 'networkidle0' });

            await page.waitForNavigation();
            console.log('now patreon login');

            await page.waitForNavigation();
            console.log('now click patreon allow');

            await page.waitForNavigation({ waitUntil: 'load' });
            console.log('now credentialed');

        }

        // temp
        //hrefs = hrefs.slice(0, 100);

        // download episodes
        let episodes = [];
        {
            for (let i = 0; i < hrefs.length; i++) {
                const url = hrefs[i];

                console.log(`Loading story ${i} ${url}...`);
                // Navigate to the selected page
                await page.goto(url, { waitUntil: 'load' });

                let title = await page.$eval('.entry-title', x => x.textContent);
                let date = await page.$eval('footer div:last-child', x => x.textContent);

                let audioUrl = '';
                let transcriptUrl = '';

                try {
                    transcriptUrl = await page.$eval('.entry-content-wrap footer div:nth-child(2) a', x => x.href);
                } catch (ex) {

                }

                try {
                    audioUrl = await page.$eval('audio', x => x.src);
                } catch (ex) {

                }

                date = date.replace('Published on ', '');

                let transcript = 'no transcript';

                try {
                    transcript = await page.$eval('div[aria-labelledby="tab-nofurigana"]', x => x.textContent);
                } catch (ex) {
                    console.log('falling back to first transcript tab');
                    try {
                        transcript = await page.$eval('div[aria-labelledby]', x => x.textContent);  // fall back
                    } catch (ex) {
                        console.log('no transcript');
                    }
                }

                episodes.push({
                    title: title.trim(),
                    date: date.trim(),
                    audioUrl: audioUrl,
                    transcriptUrl: transcriptUrl,
                    transcript: transcript.trim()
                });
            }
        }

        return episodes;
    }
}

const norikoScraperObject = {
    url: 'https://www.japanesewithnoriko.com/season-1-archive',
    async scraper(browser) {

        
        const audioFolder = '../../static/audio/Learn Japanese with Noriko - Season 1';

        let audioFilesByEpisodeNumber = {};

        const regex = /-(\d*)\./gm;
        fs.readdirSync(audioFolder).forEach(file => {
            
            let m;

            let isMatch = false;

            while ((m = regex.exec(file)) !== null) {
                // This is necessary to avoid infinite loops with zero-width matches
                if (m.index === regex.lastIndex) {
                    regex.lastIndex++;
                }

                var group = m[1];
                if (!group) {
                    console.log('could not parse episode number for: ', file);
                } else {
                    isMatch = true;
                    audioFilesByEpisodeNumber[group] = 'SAKURA TIPS｜Listen to Japanese/' + file;
                }
            }

            if (!isMatch) {
                console.log('no match: ', file);
            }
        });

       // console.log(audioFilesByEpisodeNumber);

        let page = await browser.newPage();

        let episodes = [];

        console.log(`Navigating to ${this.url}...`);
        // Navigate to the selected page
        await page.goto(this.url, { waitUntil: 'networkidle0' });

        let links = await page.$$eval('.archive-item-link', links => {
            links = links.map(el => el.href);
            return links;
        });

        for (const link of links) {
            console.log(`Episode ${link}...`);// Navigate to the selected page
            await page.goto(link, { waitUntil: 'networkidle0' });

            let components = link.split('/');

            let numAndTitle = await page.$eval('.entry-title', title => title.textContent.trim().split('.'));
            let epNumber = numAndTitle[0];
            title = numAndTitle.slice(1).join('.');
            let content = await page.$eval('.sqs-html-content', content => content.textContent.trim());
            let date = await page.$eval('time span', el => el.textContent.trim())

            let ep = {
                link: link,
                title: title,
                date: date,
                episodeNumber: epNumber,
                content: content,
                contentFormat: 'text',
                audio: audioFilesByEpisodeNumber[epNumber]
            };
            console.log(ep);

            episodes.push(ep);
        }

        return episodes;
    }
}

const sakuraScraperObject = {
    url: 'https://sakuratips.com/category/pod-cast/page/',   // base url
    async scraper(browser) {
        

        const audioFolder = '../../static/audio/SAKURA TIPS｜Listen to Japanese';

        let audioFilesByEpisodeNumber = {};

        const regex = /-(\d*)\./gm;
        fs.readdirSync(audioFolder).forEach(file => {
            
            let m;

            while ((m = regex.exec(file)) !== null) {
                // This is necessary to avoid infinite loops with zero-width matches
                if (m.index === regex.lastIndex) {
                    regex.lastIndex++;
                }

                var group = m[1];
                if (!group) {
                    console.log('could not parse episode number for: ', file);
                } else {
                    audioFilesByEpisodeNumber[group] = 'SAKURA TIPS｜Listen to Japanese/' + file;
                }
            }
        });

        //console.log(audioFilesByEpisodeNumber);

        let page = await browser.newPage();


        // e.g. to get page 3 https://sakuratips.com/category/pod-cast/page/3/
        // pages go up to 52

        // episode list: document.getElementsByClassName('post-list-meta')

        // get date and episode number from the url

        // title: document.getElementsByClassName('cps-post-title')[0].textContent.trim()

        // trasncript: document.getElementsByClassName('cps-post-main')[0].textContent.trim()


        let episodes = [];

        for (let i = 1; i <= 52; i++) {
            let url = this.url + i + '/';
            console.log(`Navigating to ${url}...`);
            // Navigate to the selected page
            await page.goto(url, { waitUntil: 'networkidle0' });

            let links = await page.$$eval('.post-list-link', links => {
                links = links.map(el => el.href);
                return links;
            });

            for (const link of links) {
                console.log(`Episode ${link}...`);// Navigate to the selected page
                await page.goto(link, { waitUntil: 'networkidle0' });

                let components = link.split('/');

                let title = await page.$eval('.cps-post-title', title => title.textContent.trim());
                let content = await page.$eval('.cps-post-main', content => content.textContent.trim());

                let epNumber = components[components.length - 2];

                console.log('adding ep with audio: ' + audioFilesByEpisodeNumber[epNumber]);

                episodes.push({
                    linke: link,
                    title: title,
                    date: components.slice(-5, -2).join('/'),
                    episodeNumber: epNumber,
                    content: content,
                    contentFormat: 'text',
                    audio: audioFilesByEpisodeNumber[epNumber]
                });
            }
        }

        // console.log(episodes);

        return episodes;
    }
};


const picnicScraperObject = {
    url: 'https://nihongopicnic.notion.site/Nihongo-Picnic-Podcast-s-Transcript-e6c923a2d9f34c1fa278bb5e4531ea0f',
    async scraper(browser) {

        const audioFolder = '../../static/audio/Nihongo Picnic Podcast';

        let audioFiles = [];

        fs.readdirSync(audioFolder).forEach(file => {
            //console.log(file);
            audioFiles.push(file);
        });

        let page = await browser.newPage();
        console.log(`Navigating to ${this.url}...`);
        // Navigate to the selected page
        await page.goto(this.url, { waitUntil: 'networkidle0' });

        for (; ;) {
            const isMore = await page.evaluate(() => {
                let eles = document.getElementsByClassName('arrowDown');
                if (eles.length == 0) {
                    return false;
                }
                eles[0].parentElement.click();
                return true;
            });
            if (!isMore) {
                break;
            }
        }

        let episodes = await page.$$eval('.notion-table-view-row', rows => {

            //return row.querySelectorAll('span').map(el => el.textContent);
            rows = rows.map(el => {
                let text = Array.from(el.querySelectorAll('.notion-table-view-cell')).map(element => {
                    return element.textContent;
                });
                let links = Array.from(el.querySelectorAll('.notion-table-view-cell a')).map(element => {
                    return element.href;
                });
                return {
                    episodeNumber: text[0],
                    title: text[1],
                    date: text[2],
                    level: text[3].trim(),
                    link: links[0]
                };
            });
            return rows;
        });
        console.log('episode count', episodes.length);

        let scrapedData = [];

        let episodePromise = (ep) => new Promise(async (resolve, reject) => {
            let newPage = await browser.newPage();
            await newPage.goto(ep.link, { waitUntil: 'networkidle0' });
            ep.content = await newPage.$eval('.notion-page-content', text => text.textContent);
            ep.contentFormat = 'text';

            let matches = [];
            let epString = ' ' + ep.episodeNumber + '. ';
            for (var f of audioFiles) {
                if (f.includes(epString)) {
                    ep.audio = 'Nihongo Picnic Podcast/' + f;
                    matches.push(f);
                }
            }
            if (matches.length != 1) {
                console.log(ep.episodeNumber, matches);
            }

            resolve(ep);
            await newPage.close();
        });

        //episodes = episodes.slice(0, 3);
        console.log('eps', episodes.length);

        for (let i = 0; i < episodes.length; i++) {
            scrapedData[i] = await episodePromise(episodes[i]);
        }

        // rss: https://anchor.fm/s/2e08a010/podcast/rss
        // `<title><![CDATA[`   <epnumber><dot> <title>`]]></title>

        // find the rss entry that has the rss file


        page.close();

        return scrapedData;
    }
};

