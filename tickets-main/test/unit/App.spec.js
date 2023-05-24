import App from "@/App";
import Navbar from "@/components/Navbar";
import { shallowMount } from "@vue/test-utils";
import { expect } from "chai";

describe("App", () => {
  const wrapper = shallowMount(App)

  it("is called App", () => {
    expect(wrapper.name()).to.equal('app');
  });

  it("should render Navbar", () => {
    expect(wrapper.find(Navbar).exists()).to.be.true;
  });
});