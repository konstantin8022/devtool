import Cards from '@/components/Cards'
import { shallowMount } from '@vue/test-utils'
import { expect } from 'chai'

describe('Cards', () => {
  it('has correct props and attributes', () => {

    const movies = [
      {
        'id': 1,
        'title': 'Star Wars',
        'imageUrl': 'seance-1.jpg'
      }
    ]

    const wrapper = shallowMount(Cards,{
      propsData: {
        movies
      }
    })

    expect(wrapper.find('.card-title').text()).to.equal('Star Wars')
    expect(wrapper.find('#movie-0').text()).to.equal('Выберите фильм')
    expect(wrapper.find('img').attributes('src')).to.equal('seance-1.jpg')

    wrapper.find('#movie-0').trigger('click')
    expect(wrapper.vm.showModal).to.equal(true);
    // console.log(wrapper.html())
  })

  it('check showModal default property', () => {
    const wrapper = shallowMount(Cards)
    expect(wrapper.vm.showModal).to.equal(false);
  })
})