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

        // // nihongo picnic
        // let scrapedData = await picnicScraperObject.scraper(browser);
        // await browser.close();
        // fs.writeFile("nihongo-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
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
        // fs.writeFile("noriko-season2-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
        //     if (err) {
        //         return console.log(err);
        //     }
        //     console.log("The data has been scraped and saved successfully! View it at './noriko-data.json'");
        // });

        // cj
        scrapedData = await cjScraperObject.scraper(browser);
        await browser.close();
        fs.writeFile("cj-data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
            if (err) {
                return console.log(err);
            }
            console.log("The data has been scraped and saved successfully! View it at './cj-data.json'");
        });

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
    url: 'https://www.japanesewithnoriko.com/season-2-transcription',
    async scraper(browser) {
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

            episodes.push({
                url: link,
                title: title,
                date: date,
                episodeNumber: epNumber,
                content: content
            });
        }

        return episodes;
    }
}

const sakuraScraperObject = {
    // to get page 3 https://sakuratips.com/category/pod-cast/page/3/
    // pages go up to 52

    // episode list: document.getElementsByClassName('post-list-meta')

    // get date and episode number from the url

    // title: document.getElementsByClassName('cps-post-title')[0].textContent.trim()

    // trasncript: document.getElementsByClassName('cps-post-main')[0].textContent.trim()


    url: 'https://sakuratips.com/category/pod-cast/page/',   // base url
    async scraper(browser) {
        let page = await browser.newPage();

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

                episodes.push({
                    url: link,
                    title: title,
                    date: components.slice(-5, -2).join('/'),
                    episodeNumber: components[components.length - 2],
                    content: content
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
                    epNumber: text[0],
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
            resolve(ep);
            await newPage.close();
        });

        //episodes = episodes.slice(0, 3);
        console.log('eps', episodes.length);

        for (let i = 0; i < episodes.length; i++) {
            scrapedData[i] = await episodePromise(episodes[i]);
        }

        page.close();

        return scrapedData;

        // const [response] = await Promise.all([
        //     page.waitForNavigation(waitOptions),
        //     page.click(selector, clickOptions),
        //   ]);



        // Wait for the required DOM to be rendered
        async function scrapeCurrentPage() {

            // await page.waitForSelector('.page_inner');
            // // Get the link to all the required books
            // let urls = await page.$$eval('section ol > li', links => {
            //     // Make sure the book to be scraped is in stock
            //     links = links.filter(link => link.querySelector('.instock.availability > i').textContent !== "In stock")
            //     // Extract the links from the data
            //     links = links.map(el => el.querySelector('h3 > a').href)
            //     return links;
            // });

            // Loop through each of those links, open a new page instance and get the relevant data from them
            let pagePromise = (link) => new Promise(async (resolve, reject) => {
                let dataObj = {};
                let newPage = await browser.newPage();
                await newPage.goto(link);
                dataObj['bookTitle'] = await newPage.$eval('.product_main > h1', text => text.textContent);
                dataObj['bookPrice'] = await newPage.$eval('.price_color', text => text.textContent);
                dataObj['noAvailable'] = await newPage.$eval('.instock.availability', text => {
                    // Strip new line and tab spaces
                    text = text.textContent.replace(/(\r\n\t|\n|\r|\t)/gm, "");
                    // Get the number of stock available
                    let regexp = /^.*\((.*)\).*$/i;
                    let stockAvailable = regexp.exec(text)[1].split(' ')[0];
                    return stockAvailable;
                });
                dataObj['imageUrl'] = await newPage.$eval('#product_gallery img', img => img.src);
                dataObj['bookDescription'] = await newPage.$eval('#product_description', div => div.nextSibling.nextSibling.textContent);
                dataObj['upc'] = await newPage.$eval('.table.table-striped > tbody > tr > td', table => table.textContent);
                resolve(dataObj);
                await newPage.close();
            });

            for (link in urls) {
                let currentPageData = await pagePromise(urls[link]);
                scrapedData.push(currentPageData);
                // console.log(currentPageData);
            }
            // When all the data on this page is done, click the next button and start the scraping of the next page
            // You are going to check if this button exist first, so you know if there really is a next page.
            let nextButtonExist = false;
            try {
                const nextButton = await page.$eval('.next > a', a => a.textContent);
                nextButtonExist = true;
            }
            catch (err) {
                nextButtonExist = false;
            }
            if (nextButtonExist) {
                await page.click('.next > a');
                return scrapeCurrentPage(); // Call this function recursively
            }
            await page.close();
            return scrapedData;
        }
        // let data = await scrapeCurrentPage();
        // console.log(data);
        data = 'foo'
        return data;
    }
};

