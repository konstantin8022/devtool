require('chromedriver');
const assert = require('assert');
const {Builder, Key, By, until} = require('selenium-webdriver');

describe('Checkout booking', function () {
  let driver;

  before(async function() {
    driver = await new Builder().forBrowser('chrome').build();
  });

  it('success', async function() {
    await driver.get('http://localhost:8080/');

    await driver.sleep(500);

    await driver.wait(until.elementLocated(By.id('cities-dropdown')), 5000);
    await driver.findElement(By.id('cities-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('city-1')), 5000);
    await driver.findElement(By.id('city-1')).click();

    await driver.wait(until.elementLocated(By.id('movie-1')), 10000);
    await driver.findElement(By.id('movie-1')).click();

    await driver.wait(until.elementLocated(By.id('seances-dropdown')), 5000);
    await driver.findElement(By.id('seances-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('seance-2')), 5000);
    await driver.findElement(By.id('seance-2')).click();

    await driver.wait(until.elementLocated(By.id('cinema-hall')), 5000);

    await driver.wait(until.elementLocated(By.className('seat')), 5000);
    await driver.findElement(By.className('seat')).click();

    await driver.findElement(By.id('email-field')).click();
    await driver.findElement(By.id('email-field')).sendKeys('test@mail.com');

    await driver.findElement(By.id('submit-button')).click();

    await driver.wait(until.elementLocated(By.className('b-toast-success')), 5000);

    await driver.sleep(500);

    let text = await driver.findElement(By.className('toast-body')).getText();

    assert.equal(text, 'Заказ был успешно выполнен');
  });
  
  after(() => driver && driver.quit());
})